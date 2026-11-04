package http

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tiquetera/internal/auth"
	"tiquetera/internal/domain"
)

const testConsumoQR = "TQ1.qr-secreto"

type fakeConsumos struct {
	err    error
	cajero uuid.UUID
	key    uuid.UUID
	qr     string
	calls  int
}

func (f *fakeConsumos) Consume(_ context.Context, cajero, key uuid.UUID, qr string) (domain.ConsumoResult, error) {
	f.cajero, f.key, f.qr = cajero, key, qr
	f.calls++
	return domain.ConsumoResult{
		ClienteNombre:    "Ana P.",
		SaldoRestante:    7,
		FechaVencimiento: time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		FechaHora:        time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC),
	}, f.err
}

type consumosEnv struct {
	router http.Handler
	svc    *fakeConsumos
	tokens *auth.TokenManager
	logs   *bytes.Buffer
}

func newConsumosEnv(t *testing.T) consumosEnv {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	tokens := auth.NewTokenManager(key, &stubClock{t: time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)})
	svc := &fakeConsumos{}
	logs := &bytes.Buffer{}
	router := NewRouter(Deps{
		Log: slog.New(slog.NewJSONHandler(logs, nil)), DB: fakePinger{},
		Tokens: tokens, Auth: &fakeAuth{}, Consumos: svc,
	})
	return consumosEnv{router: router, svc: svc, tokens: tokens, logs: logs}
}

func (e consumosEnv) post(t *testing.T, user uuid.UUID, rol domain.Role, key, body string) *httptest.ResponseRecorder {
	t.Helper()
	token, err := e.tokens.IssueAccess(domain.Principal{UserID: user, Rol: rol})
	require.NoError(t, err)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/consumos", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	return rec
}

func validConsumoBody() string { return `{"qr":"` + testConsumoQR + `"}` }

func TestConsumos_Success_ReturnsContract(t *testing.T) {
	for _, rol := range []domain.Role{domain.RoleCajero, domain.RoleAdmin} {
		t.Run(string(rol), func(t *testing.T) {
			e := newConsumosEnv(t)
			user, key := uuid.New(), uuid.New()
			rec := e.post(t, user, rol, key.String(), validConsumoBody())

			require.Equal(t, http.StatusOK, rec.Code)
			assert.JSONEq(t, `{"ok":true,"cliente":{"nombre":"Ana P."},"saldo_restante":7,"fecha_vencimiento":"2026-12-31","fecha_hora":"2026-11-04T12:00:00Z"}`, rec.Body.String())
			assert.Equal(t, user, e.svc.cajero)
			assert.Equal(t, key, e.svc.key)
			assert.Equal(t, testConsumoQR, e.svc.qr)
			assert.NotContains(t, e.logs.String(), testConsumoQR)
		})
	}
}

func TestConsumos_WithoutToken_Returns401(t *testing.T) {
	e := newConsumosEnv(t)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/consumos", strings.NewReader(validConsumoBody()))
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestConsumos_InvalidRequest_Returns400WithoutCallingService(t *testing.T) {
	tests := []struct{ name, key, body string }{
		{"missing key", "", validConsumoBody()},
		{"malformed key", "abc", validConsumoBody()},
		{"nil key", uuid.Nil.String(), validConsumoBody()},
		{"not json", uuid.NewString(), "nope"},
		{"unknown field", uuid.NewString(), `{"qr":"x","saldo":9}`},
		{"empty qr", uuid.NewString(), `{"qr":""}`},
		{"missing qr", uuid.NewString(), `{}`},
		{"long qr", uuid.NewString(), `{"qr":"` + strings.Repeat("a", 129) + `"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newConsumosEnv(t)
			rec := e.post(t, uuid.New(), domain.RoleCajero, tt.key, tt.body)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), "INVALID_REQUEST")
			assert.Zero(t, e.svc.calls)
		})
	}
}

func TestConsumos_Errors_AreMappedToProblems(t *testing.T) {
	expired := domain.ConsumoRejection{Code: domain.RejectTiqueteraExpired, Message: "Tiquetera vencida el 30/10/2026"}
	tests := []struct {
		name   string
		err    error
		status int
		want   string
	}{
		{"expired keeps message", expired, http.StatusUnprocessableEntity, "Tiquetera vencida el 30/10/2026"},
		{"no balance", domain.NewConsumoRejection(domain.RejectNoBalance), http.StatusUnprocessableEntity, "SIN_SALDO"},
		{"bad qr", domain.NewConsumoRejection(domain.RejectQRInvalid), http.StatusUnprocessableEntity, "QR_INVALIDO"},
		{"unknown qr", domain.NewConsumoRejection(domain.RejectClienteNotFound), http.StatusUnprocessableEntity, "CLIENTE_NO_EXISTE"},
		{"inactive", domain.NewConsumoRejection(domain.RejectClienteInactive), http.StatusUnprocessableEntity, "CLIENTE_INACTIVO"},
		{"no book", domain.NewConsumoRejection(domain.RejectNoTiquetera), http.StatusUnprocessableEntity, "SIN_TIQUETERA"},
		{"foreign key", domain.ErrIdempotencyConflict, http.StatusConflict, "IDEMPOTENCY_KEY_CONFLICT"},
		{"unexpected", errors.New("detalle interno"), http.StatusInternalServerError, "INTERNAL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newConsumosEnv(t)
			e.svc.err = tt.err
			rec := e.post(t, uuid.New(), domain.RoleCajero, uuid.NewString(), validConsumoBody())
			assert.Equal(t, tt.status, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.want)
			assert.NotContains(t, rec.Body.String(), "detalle interno")
			assert.NotContains(t, rec.Body.String(), `"ok"`)
		})
	}
}

func TestConsumos_RateLimit_IsPerCajero(t *testing.T) {
	e := newConsumosEnv(t)
	busy, other := uuid.New(), uuid.New()
	for range consumoRateLimit {
		require.Equal(t, http.StatusOK, e.post(t, busy, domain.RoleCajero, uuid.NewString(), validConsumoBody()).Code)
	}

	limited := e.post(t, busy, domain.RoleCajero, uuid.NewString(), validConsumoBody())
	assert.Equal(t, http.StatusTooManyRequests, limited.Code)
	assert.Contains(t, limited.Body.String(), "RATE_LIMITED")
	assert.Equal(t, consumoRateLimit, e.svc.calls)

	assert.Equal(t, http.StatusOK, e.post(t, other, domain.RoleCajero, uuid.NewString(), validConsumoBody()).Code)
}
