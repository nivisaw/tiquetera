package http

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
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
	"tiquetera/internal/service"
)

const (
	testQRPayload   = "TQ1.token-secreto"
	validClienteReq = `{"nombre":"Ana","telefono":"3001234567","almuerzos":10,"fecha_vencimiento":"2026-12-01"}`
	validRecargaReq = `{"almuerzos":5,"fecha_vencimiento":"2026-12-01"}`
)

type fakeClientes struct {
	err       error
	cliente   domain.Cliente
	tiquetera domain.Tiquetera
	noActive  bool
	patch     domain.ClientePatch
	page      int32
	size      int32
	q         string
	actor     uuid.UUID
}

func (f *fakeClientes) Create(_ context.Context, actor uuid.UUID, _ domain.ClienteData, _ domain.TiqueteraInput) (service.ClienteCreated, error) {
	f.actor = actor
	return service.ClienteCreated{Cliente: f.cliente, Tiquetera: f.tiquetera, QRPayload: testQRPayload}, f.err
}

func (f *fakeClientes) Get(context.Context, uuid.UUID) (service.ClienteDetail, error) {
	detail := service.ClienteDetail{Cliente: f.cliente}
	if !f.noActive {
		detail.Tiquetera = &f.tiquetera
	}
	return detail, f.err
}

func (f *fakeClientes) List(_ context.Context, q string, page, size int32) ([]domain.Cliente, int64, error) {
	f.q, f.page, f.size = q, page, size
	return []domain.Cliente{f.cliente}, 1, f.err
}

func (f *fakeClientes) Update(_ context.Context, _, _ uuid.UUID, patch domain.ClientePatch) (domain.Cliente, error) {
	f.patch = patch
	return f.cliente, f.err
}

func (f *fakeClientes) Recharge(context.Context, uuid.UUID, uuid.UUID, domain.TiqueteraInput) (domain.Tiquetera, error) {
	return f.tiquetera, f.err
}

func (f *fakeClientes) RegenerateQR(context.Context, uuid.UUID, uuid.UUID) (string, error) {
	return testQRPayload, f.err
}

type clientesEnv struct {
	router http.Handler
	svc    *fakeClientes
	tokens *auth.TokenManager
	logs   *bytes.Buffer
	userID uuid.UUID
}

func newClientesEnv(t *testing.T) clientesEnv {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	tokens := auth.NewTokenManager(key, &stubClock{t: time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)})
	svc := &fakeClientes{
		cliente:   domain.Cliente{ClienteData: domain.ClienteData{Nombre: "Ana", Telefono: "+573001234567"}, ID: uuid.New(), Activo: true, QRVersion: 1},
		tiquetera: domain.Tiquetera{ID: uuid.New(), Saldo: 5, FechaVencimiento: time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC), Activa: true},
	}
	logs := &bytes.Buffer{}
	router := NewRouter(Deps{
		Log: slog.New(slog.NewJSONHandler(logs, nil)), DB: fakePinger{},
		Tokens: tokens, Auth: &fakeAuth{}, Clientes: svc,
	})
	return clientesEnv{router: router, svc: svc, tokens: tokens, logs: logs, userID: uuid.New()}
}

func (e clientesEnv) call(t *testing.T, rol domain.Role, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	token, err := e.tokens.IssueAccess(domain.Principal{UserID: e.userID, Rol: rol})
	require.NoError(t, err)
	req := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	return rec
}

func clientePaths() []struct{ method, path, body string } {
	id := uuid.NewString()
	return []struct{ method, path, body string }{
		{http.MethodPost, "/v1/clientes", validClienteReq},
		{http.MethodGet, "/v1/clientes", ""},
		{http.MethodGet, "/v1/clientes/" + id, ""},
		{http.MethodPatch, "/v1/clientes/" + id, `{"nombre":"Rosa"}`},
		{http.MethodPost, "/v1/clientes/" + id + "/tiqueteras", validRecargaReq},
		{http.MethodPost, "/v1/clientes/" + id + "/qr/regenerar", ""},
	}
}

func TestClientes_EveryRoute_RejectsCajeroAndAnonymous(t *testing.T) {
	e := newClientesEnv(t)
	for _, c := range clientePaths() {
		t.Run(c.method+" "+c.path, func(t *testing.T) {
			assert.Equal(t, http.StatusForbidden, e.call(t, domain.RoleCajero, c.method, c.path, c.body).Code)
			req := httptest.NewRequestWithContext(t.Context(), c.method, c.path, strings.NewReader(c.body))
			rec := httptest.NewRecorder()
			e.router.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	}
}

func TestClientes_DomainErrors_AreMappedToProblems(t *testing.T) {
	tests := []struct {
		err    error
		status int
		code   string
	}{
		{domain.ValidationError{Message: "El teléfono no es válido"}, http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		{fmt.Errorf("wrapped: %w", domain.ErrPhoneTaken), http.StatusConflict, "TELEFONO_DUPLICADO"},
		{fmt.Errorf("wrapped: %w", domain.ErrNotFound), http.StatusNotFound, "NOT_FOUND"},
		{domain.ErrClienteInactive, http.StatusUnprocessableEntity, "CLIENTE_INACTIVO"},
		{errors.New("db secret detail"), http.StatusInternalServerError, "INTERNAL"},
	}
	for _, tt := range tests {
		e := newClientesEnv(t)
		e.svc.err = tt.err
		for _, c := range clientePaths() {
			t.Run(tt.code+" "+c.method+" "+c.path, func(t *testing.T) {
				rec := e.call(t, domain.RoleAdmin, c.method, c.path, c.body)
				assert.Equal(t, tt.status, rec.Code)
				assert.Contains(t, rec.Body.String(), tt.code)
				assert.NotContains(t, rec.Body.String(), "db secret detail")
				assert.NotContains(t, rec.Body.String(), testQRPayload)
			})
		}
	}
}

func TestClientesCreate_Valid_ReturnsPayloadAndUsesTokenActor(t *testing.T) {
	e := newClientesEnv(t)
	rec := e.call(t, domain.RoleAdmin, http.MethodPost, "/v1/clientes", validClienteReq)
	require.Equal(t, http.StatusCreated, rec.Code)

	var got createClienteResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, testQRPayload, got.QRPayload)
	assert.Equal(t, "2026-12-01", got.Tiquetera.FechaVencimiento)
	assert.Equal(t, e.userID, e.svc.actor)
	assert.NotContains(t, e.logs.String(), testQRPayload)
}

func TestClientes_MalformedBody_Returns400(t *testing.T) {
	e := newClientesEnv(t)
	id := uuid.NewString()
	tests := []struct{ name, method, path, body string }{
		{"create not json", http.MethodPost, "/v1/clientes", `nope`},
		{"create unknown field", http.MethodPost, "/v1/clientes", `{"nombre":"Ana","rol":"admin"}`},
		{"create wrong type", http.MethodPost, "/v1/clientes", `{"almuerzos":"10"}`},
		{"recharge not json", http.MethodPost, "/v1/clientes/" + id + "/tiqueteras", `nope`},
		{"patch empty", http.MethodPatch, "/v1/clientes/" + id, `{}`},
		{"patch null name", http.MethodPatch, "/v1/clientes/" + id, `{"nombre":null}`},
		{"patch null phone", http.MethodPatch, "/v1/clientes/" + id, `{"telefono":null}`},
		{"patch null active", http.MethodPatch, "/v1/clientes/" + id, `{"activo":null}`},
		{"patch wrong type", http.MethodPatch, "/v1/clientes/" + id, `{"activo":"no"}`},
		{"patch not json", http.MethodPatch, "/v1/clientes/" + id, `nope`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := e.call(t, domain.RoleAdmin, tt.method, tt.path, tt.body)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), "INVALID_REQUEST")
		})
	}
}

func TestClientesUpdate_NullClearsOptionalAbsentKeeps(t *testing.T) {
	e := newClientesEnv(t)
	path := "/v1/clientes/" + uuid.NewString()

	require.Equal(t, http.StatusOK, e.call(t, domain.RoleAdmin, http.MethodPatch, path, `{"documento":null,"activo":false}`).Code)
	assert.True(t, e.svc.patch.SetDocumento)
	assert.Nil(t, e.svc.patch.Documento)
	assert.False(t, e.svc.patch.SetEmail)
	assert.False(t, *e.svc.patch.Activo)
}

func TestClientes_MalformedID_Returns404(t *testing.T) {
	e := newClientesEnv(t)
	for _, c := range []struct{ method, path, body string }{
		{http.MethodGet, "/v1/clientes/123", ""},
		{http.MethodPatch, "/v1/clientes/123", `{"nombre":"Rosa"}`},
		{http.MethodPost, "/v1/clientes/123/tiqueteras", validRecargaReq},
		{http.MethodPost, "/v1/clientes/123/qr/regenerar", ""},
	} {
		assert.Equal(t, http.StatusNotFound, e.call(t, domain.RoleAdmin, c.method, c.path, c.body).Code, c.path)
	}
}

func TestClientesList_Pagination(t *testing.T) {
	tests := []struct {
		name, query string
		status      int
		page, size  int32
	}{
		{"defaults", "", http.StatusOK, 1, 20},
		{"explicit", "?page=3&size=50&q=%20ana%20", http.StatusOK, 3, 50},
		{"size over limit", "?size=51", http.StatusBadRequest, 0, 0},
		{"size zero", "?size=0", http.StatusBadRequest, 0, 0},
		{"page zero", "?page=0", http.StatusBadRequest, 0, 0},
		{"page too big", "?page=1000001", http.StatusBadRequest, 0, 0},
		{"not a number", "?page=x", http.StatusBadRequest, 0, 0},
		{"long query", "?q=" + strings.Repeat("a", 121), http.StatusBadRequest, 0, 0},
		{"nul in query", "?q=a%00b", http.StatusBadRequest, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newClientesEnv(t)
			rec := e.call(t, domain.RoleAdmin, http.MethodGet, "/v1/clientes"+tt.query, "")
			require.Equal(t, tt.status, rec.Code)
			if tt.status != http.StatusOK {
				return
			}
			assert.Equal(t, tt.page, e.svc.page)
			assert.Equal(t, tt.size, e.svc.size)
			var got clientePage
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
			assert.Len(t, got.Items, 1)
			assert.Equal(t, int64(1), got.Total)
		})
	}
}

func TestClientesRead_NeverExposesQR(t *testing.T) {
	e := newClientesEnv(t)
	id := uuid.NewString()
	for _, path := range []string{"/v1/clientes", "/v1/clientes/" + id} {
		rec := e.call(t, domain.RoleAdmin, http.MethodGet, path, "")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "qr_payload")
		assert.NotContains(t, rec.Body.String(), "qr_token")
		assert.Contains(t, rec.Body.String(), "qr_version")
	}
}

func TestClientesGet_ActiveBookIsNullWhenMissing(t *testing.T) {
	e := newClientesEnv(t)
	path := "/v1/clientes/" + uuid.NewString()

	assert.Contains(t, e.call(t, domain.RoleAdmin, http.MethodGet, path, "").Body.String(), `"saldo":5`)

	e.svc.noActive = true
	assert.Contains(t, e.call(t, domain.RoleAdmin, http.MethodGet, path, "").Body.String(), `"tiquetera":null`)
}

func TestClientesRegenerateQR_ReturnsPayloadWithoutLoggingIt(t *testing.T) {
	e := newClientesEnv(t)
	rec := e.call(t, domain.RoleAdmin, http.MethodPost, "/v1/clientes/"+uuid.NewString()+"/qr/regenerar", "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), testQRPayload)
	assert.NotContains(t, e.logs.String(), testQRPayload)
}

func TestClientesRecharge_Valid_Returns201(t *testing.T) {
	e := newClientesEnv(t)
	rec := e.call(t, domain.RoleAdmin, http.MethodPost, "/v1/clientes/"+uuid.NewString()+"/tiqueteras", validRecargaReq)
	assert.Equal(t, http.StatusCreated, rec.Code)
}
