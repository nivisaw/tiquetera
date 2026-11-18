package http

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tiquetera/internal/auth"
	"tiquetera/internal/domain"
)

type fakeReportes struct {
	err          error
	param        int32
	page, size   int32
	desde, hasta time.Time
	cajero       *uuid.UUID
}

var reporteFecha = time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)

func (f *fakeReportes) Dashboard(context.Context) (domain.Dashboard, error) {
	return domain.Dashboard{
		ConsumosHoy: 3,
		Consumos7d:  []domain.DiaTotal{{Fecha: reporteFecha, Total: 3}},
		Vigentes:    domain.ResumenVigentes{Clientes: 4, SaldoPendiente: 40, PorVencer: 1},
		VendidosMes: 60,
		RechazosHoy: 2,
	}, f.err
}

func (f *fakeReportes) list(param, page, size int32) ([]domain.ClienteReporte, int64, error) {
	f.param, f.page, f.size = param, page, size
	return []domain.ClienteReporte{{ClienteID: uuid.New(), Nombre: "Ana", Telefono: "+573001234567", Fecha: reporteFecha, Saldo: 2}}, 1, f.err
}

func (f *fakeReportes) SinRecarga(_ context.Context, dias, page, size int32) ([]domain.ClienteReporte, int64, error) {
	return f.list(dias, page, size)
}

func (f *fakeReportes) PorVencer(_ context.Context, dias, page, size int32) ([]domain.ClienteReporte, int64, error) {
	return f.list(dias, page, size)
}

func (f *fakeReportes) SaldoBajo(_ context.Context, maxSaldo, page, size int32) ([]domain.ClienteReporte, int64, error) {
	return f.list(maxSaldo, page, size)
}

func (f *fakeReportes) Consumos(_ context.Context, desde, hasta time.Time, cajero *uuid.UUID) (domain.ConsumosReporte, error) {
	f.desde, f.hasta, f.cajero = desde, hasta, cajero
	if f.err == nil {
		if err := domain.ValidateRango(desde, hasta); err != nil {
			return domain.ConsumosReporte{}, err
		}
	}
	return domain.ConsumosReporte{
		Total:   3,
		Dias:    []domain.DiaTotal{{Fecha: reporteFecha, Total: 3}},
		Cajeros: []domain.CajeroTotal{{CajeroID: uuid.New(), Nombre: "Caja 1", Total: 3}},
	}, f.err
}

type reportesEnv struct {
	clientesEnv

	rep *fakeReportes
}

func newReportesEnv(t *testing.T) reportesEnv {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	tokens := auth.NewTokenManager(key, &stubClock{t: time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)})
	rep := &fakeReportes{}
	router := NewRouter(Deps{
		Log: slog.New(slog.NewJSONHandler(io.Discard, nil)), DB: fakePinger{},
		Tokens: tokens, Auth: &fakeAuth{}, Reportes: rep,
	})
	return reportesEnv{clientesEnv: clientesEnv{router: router, tokens: tokens, userID: uuid.New()}, rep: rep}
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var out T
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	return out
}

var reportePaths = []string{
	"/v1/reportes/dashboard",
	"/v1/reportes/sin-recarga",
	"/v1/reportes/por-vencer",
	"/v1/reportes/saldo-bajo",
	"/v1/reportes/consumos?desde=2026-11-01&hasta=2026-11-04",
}

func TestReportes_EveryRoute_RejectsCajeroAndAnonymous(t *testing.T) {
	e := newReportesEnv(t)
	for _, path := range reportePaths {
		t.Run(path, func(t *testing.T) {
			assert.Equal(t, http.StatusForbidden, e.call(t, domain.RoleCajero, http.MethodGet, path, "").Code)
			rec := httptest.NewRecorder()
			e.router.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.Equal(t, http.StatusOK, e.call(t, domain.RoleAdmin, http.MethodGet, path, "").Code)
		})
	}
}

func TestReportesDashboard_Admin_ReturnsFigures(t *testing.T) {
	e := newReportesEnv(t)
	rec := e.call(t, domain.RoleAdmin, http.MethodGet, "/v1/reportes/dashboard", "")
	require.Equal(t, http.StatusOK, rec.Code)
	got := decode[dashboardDTO](t, rec)
	assert.Equal(t, dashboardDTO{
		ConsumosHoy: 3, Consumos7d: []diaTotalDTO{{Fecha: "2026-11-04", Total: 3}},
		ClientesVigentes: 4, SaldoPendiente: 40, VendidosMes: 60, PorVencer7d: 1, RechazosHoy: 2,
	}, got)
}

func TestReportesLists_DefaultsAndParams(t *testing.T) {
	tests := []struct {
		path        string
		param, size int32
		field       string
	}{
		{"/v1/reportes/sin-recarga", 30, 20, "ultima_recarga"},
		{"/v1/reportes/sin-recarga?dias=60&page=2&size=50", 60, 50, "ultima_recarga"},
		{"/v1/reportes/por-vencer", 7, 20, "fecha_vencimiento"},
		{"/v1/reportes/por-vencer?dias=0", 0, 20, "fecha_vencimiento"},
		{"/v1/reportes/saldo-bajo", 3, 20, "fecha_vencimiento"},
		{"/v1/reportes/saldo-bajo?max=0", 0, 20, "fecha_vencimiento"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			e := newReportesEnv(t)
			rec := e.call(t, domain.RoleAdmin, http.MethodGet, tt.path, "")
			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tt.param, e.rep.param)
			assert.Equal(t, tt.size, e.rep.size)
			body := decode[map[string]any](t, rec)
			items, ok := body["items"].([]any)
			require.True(t, ok)
			require.Len(t, items, 1)
			assert.Contains(t, items[0], tt.field)
			assert.InDelta(t, 1, body["total"], 0)
		})
	}
}

func TestReportes_InvalidParams_Return422(t *testing.T) {
	paths := []string{
		"/v1/reportes/sin-recarga?dias=0",
		"/v1/reportes/sin-recarga?dias=366",
		"/v1/reportes/por-vencer?dias=61",
		"/v1/reportes/por-vencer?dias=x",
		"/v1/reportes/saldo-bajo?max=-1",
		"/v1/reportes/saldo-bajo?max=51",
		"/v1/reportes/consumos",
		"/v1/reportes/consumos?desde=2026-11-01",
		"/v1/reportes/consumos?desde=2026-13-01&hasta=2026-11-04",
		"/v1/reportes/consumos?desde=2026-11-04&hasta=2026-11-01",
		"/v1/reportes/consumos?desde=2026-10-01&hasta=2026-11-01",
		"/v1/reportes/consumos?desde=2026-11-01&hasta=2026-11-04&cajero_id=x",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			e := newReportesEnv(t)
			rec := e.call(t, domain.RoleAdmin, http.MethodGet, path, "")
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Contains(t, rec.Body.String(), "VALIDATION_ERROR")
		})
	}
}

func TestReportes_InvalidPagination_Returns400(t *testing.T) {
	e := newReportesEnv(t)
	rec := e.call(t, domain.RoleAdmin, http.MethodGet, "/v1/reportes/saldo-bajo?size=51", "")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestReportesConsumos_ParsesRangeAndCashier(t *testing.T) {
	e := newReportesEnv(t)
	cajero := uuid.New()
	rec := e.call(t, domain.RoleAdmin, http.MethodGet,
		"/v1/reportes/consumos?desde=2026-10-05&hasta=2026-11-04&cajero_id="+cajero.String(), "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC), e.rep.desde)
	assert.Equal(t, reporteFecha, e.rep.hasta)
	require.NotNil(t, e.rep.cajero)
	assert.Equal(t, cajero, *e.rep.cajero)
	got := decode[consumosReporteDTO](t, rec)
	assert.Equal(t, int64(3), got.Total)
	assert.Equal(t, "Caja 1", got.Cajeros[0].Nombre)
}

func TestReportes_ServiceError_Returns500Generic(t *testing.T) {
	for _, path := range reportePaths {
		t.Run(path, func(t *testing.T) {
			e := newReportesEnv(t)
			e.rep.err = errors.New("db down")
			rec := e.call(t, domain.RoleAdmin, http.MethodGet, path, "")
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.NotContains(t, rec.Body.String(), "db down")
		})
	}
}
