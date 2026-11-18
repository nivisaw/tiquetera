package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tiquetera/internal/domain"
)

var errReporteDB = errors.New("db down")

type fakeReporteRepo struct {
	failOn string
	args   map[string][]time.Time
	limit  int32
	offset int32
	porDia []domain.DiaTotal
}

func (f *fakeReporteRepo) record(name string, ts ...time.Time) error {
	if f.args == nil {
		f.args = map[string][]time.Time{}
	}
	f.args[name] = ts
	if f.failOn == name {
		return errReporteDB
	}
	return nil
}

func (f *fakeReporteRepo) CountRechazos(_ context.Context, desde, hasta time.Time) (int64, error) {
	return 2, f.record("CountRechazos", desde, hasta)
}

func (f *fakeReporteRepo) ConsumosPorDia(_ context.Context, desde, hasta time.Time, _ *uuid.UUID) ([]domain.DiaTotal, error) {
	return f.porDia, f.record("ConsumosPorDia", desde, hasta)
}

func (f *fakeReporteRepo) ConsumosPorCajero(_ context.Context, desde, hasta time.Time, _ *uuid.UUID) ([]domain.CajeroTotal, error) {
	return []domain.CajeroTotal{{Nombre: "Caja", Total: 5}}, f.record("ConsumosPorCajero", desde, hasta)
}

func (f *fakeReporteRepo) ResumenVigentes(_ context.Context, hoy, limite time.Time) (domain.ResumenVigentes, error) {
	return domain.ResumenVigentes{Clientes: 1}, f.record("ResumenVigentes", hoy, limite)
}

func (f *fakeReporteRepo) VendidosDesde(_ context.Context, desde time.Time) (int64, error) {
	return 9, f.record("VendidosDesde", desde)
}

func (f *fakeReporteRepo) page(name string, limit, offset int32, ts ...time.Time) ([]domain.ClienteReporte, int64, error) {
	f.limit, f.offset = limit, offset
	return []domain.ClienteReporte{{Nombre: "Ana"}}, 1, f.record(name, ts...)
}

func (f *fakeReporteRepo) SinRecarga(_ context.Context, corte time.Time, limit, offset int32) ([]domain.ClienteReporte, int64, error) {
	return f.page("SinRecarga", limit, offset, corte)
}

func (f *fakeReporteRepo) PorVencer(_ context.Context, hoy, limite time.Time, limit, offset int32) ([]domain.ClienteReporte, int64, error) {
	return f.page("PorVencer", limit, offset, hoy, limite)
}

func (f *fakeReporteRepo) SaldoBajo(_ context.Context, hoy time.Time, _, limit, offset int32) ([]domain.ClienteReporte, int64, error) {
	return f.page("SaldoBajo", limit, offset, hoy)
}

// 22:00 on 2026-11-04 in Bogotá, already the 5th in UTC.
func newReporteSvc(repo *fakeReporteRepo) *ReporteService {
	return NewReporteService(repo, &fakeClock{t: time.Date(2026, 11, 5, 3, 0, 0, 0, time.UTC)})
}

func day(s string) time.Time {
	d, _ := time.Parse(time.DateOnly, s)
	return d
}

func at(s string) time.Time {
	d, _ := time.Parse(time.RFC3339, s)
	return d
}

func TestReporteDashboard_UsesBogotaTodayRanges(t *testing.T) {
	repo := &fakeReporteRepo{porDia: []domain.DiaTotal{{Fecha: day("2026-11-04"), Total: 6}}}

	d, err := newReporteSvc(repo).Dashboard(t.Context())

	require.NoError(t, err)
	assert.Equal(t, int64(6), d.ConsumosHoy)
	assert.Len(t, d.Consumos7d, 7)
	assert.Equal(t, int64(2), d.RechazosHoy)
	assert.Equal(t, int64(9), d.VendidosMes)
	assert.Equal(t, []time.Time{at("2026-10-29T05:00:00Z"), at("2026-11-05T05:00:00Z")}, repo.args["ConsumosPorDia"])
	assert.Equal(t, []time.Time{at("2026-11-04T05:00:00Z"), at("2026-11-05T05:00:00Z")}, repo.args["CountRechazos"])
	assert.Equal(t, []time.Time{day("2026-11-04"), day("2026-11-11")}, repo.args["ResumenVigentes"])
	assert.Equal(t, []time.Time{at("2026-11-01T05:00:00Z")}, repo.args["VendidosDesde"])
}

func TestReporteDashboard_RepoError_IsWrapped(t *testing.T) {
	for _, step := range []string{"ConsumosPorDia", "CountRechazos", "ResumenVigentes", "VendidosDesde"} {
		t.Run(step, func(t *testing.T) {
			_, err := newReporteSvc(&fakeReporteRepo{failOn: step}).Dashboard(t.Context())
			assert.ErrorIs(t, err, errReporteDB)
		})
	}
}

func TestReporteLists_ComputeCutoffsAndOffset(t *testing.T) {
	repo := &fakeReporteRepo{}
	svc := newReporteSvc(repo)

	_, _, err := svc.SinRecarga(t.Context(), 30, 3, 20)
	require.NoError(t, err)
	assert.Equal(t, []time.Time{at("2026-10-06T05:00:00Z")}, repo.args["SinRecarga"])
	assert.Equal(t, int32(20), repo.limit)
	assert.Equal(t, int32(40), repo.offset)

	_, _, err = svc.PorVencer(t.Context(), 7, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, []time.Time{day("2026-11-04"), day("2026-11-11")}, repo.args["PorVencer"])

	_, _, err = svc.SaldoBajo(t.Context(), 3, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, []time.Time{day("2026-11-04")}, repo.args["SaldoBajo"])
}

func TestReporteLists_RepoError_IsWrapped(t *testing.T) {
	svc := newReporteSvc(&fakeReporteRepo{failOn: "SinRecarga"})
	_, _, err := svc.SinRecarga(t.Context(), 30, 1, 20)
	assert.ErrorIs(t, err, errReporteDB)

	svc = newReporteSvc(&fakeReporteRepo{failOn: "PorVencer"})
	_, _, err = svc.PorVencer(t.Context(), 7, 1, 20)
	assert.ErrorIs(t, err, errReporteDB)

	svc = newReporteSvc(&fakeReporteRepo{failOn: "SaldoBajo"})
	_, _, err = svc.SaldoBajo(t.Context(), 3, 1, 20)
	assert.ErrorIs(t, err, errReporteDB)
}

func TestReporteConsumos_FillsDaysAndTotals(t *testing.T) {
	repo := &fakeReporteRepo{porDia: []domain.DiaTotal{{Fecha: day("2026-11-02"), Total: 5}}}

	r, err := newReporteSvc(repo).Consumos(t.Context(), day("2026-11-01"), day("2026-11-03"), nil)

	require.NoError(t, err)
	assert.Equal(t, int64(5), r.Total)
	assert.Len(t, r.Dias, 3)
	assert.Len(t, r.Cajeros, 1)
	assert.Equal(t, []time.Time{at("2026-11-01T05:00:00Z"), at("2026-11-04T05:00:00Z")}, repo.args["ConsumosPorDia"])
}

func TestReporteConsumos_InvalidRangeOrRepoError(t *testing.T) {
	_, err := newReporteSvc(&fakeReporteRepo{}).Consumos(t.Context(), day("2026-11-03"), day("2026-11-01"), nil)
	var v domain.ValidationError
	require.ErrorAs(t, err, &v)

	for _, step := range []string{"ConsumosPorDia", "ConsumosPorCajero"} {
		_, err = newReporteSvc(&fakeReporteRepo{failOn: step}).Consumos(t.Context(), day("2026-11-01"), day("2026-11-01"), nil)
		assert.ErrorIs(t, err, errReporteDB)
	}
}
