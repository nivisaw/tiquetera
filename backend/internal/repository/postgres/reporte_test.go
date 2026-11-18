package postgres

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tiquetera/internal/domain"
	"tiquetera/internal/service"
)

// reporteFixture's clock reads 22:00 on 2026-11-04 in Bogotá (already the 5th
// in UTC), so "today" is 2026-11-04.
type reporteFixture struct {
	clienteFixture

	reportes *service.ReporteService
	n        int
}

func newReporteFixture(t *testing.T) *reporteFixture {
	t.Helper()
	f := newClienteFixture(t)
	_, err := f.owner.Exec(t.Context(), `UPDATE tiqueteras SET activa = false, created_at = '2026-01-01' WHERE id = $1`, f.tiquetera)
	require.NoError(t, err)
	return &reporteFixture{clienteFixture: f, reportes: service.NewReporteService(NewReporteRepository(f.pool), f.clock)}
}

// client creates a client whose active ticket book has the given balance,
// expiry and creation instant.
func (f *reporteFixture) client(t *testing.T, saldo int, vence string, createdAt time.Time) service.ClienteCreated {
	t.Helper()
	f.n++
	created, err := f.svc.Create(t.Context(), f.usuarioID,
		domain.ClienteData{Nombre: "Cliente Prueba", Telefono: fmt.Sprintf("30011122%02d", f.n)},
		domain.TiqueteraInput{Almuerzos: 20, FechaVencimiento: clienteTestDate})
	require.NoError(t, err)
	_, err = f.owner.Exec(t.Context(),
		`UPDATE tiqueteras SET saldo = $1, fecha_vencimiento = $2::date, created_at = $3 WHERE cliente_id = $4`,
		saldo, vence, createdAt, created.Cliente.ID)
	require.NoError(t, err)
	return created
}

func (f *reporteFixture) consumoAt(t *testing.T, c service.ClienteCreated, at time.Time) {
	t.Helper()
	_, err := f.owner.Exec(t.Context(),
		`INSERT INTO consumos (tiquetera_id, cliente_id, cajero_id, saldo_resultante, idempotency_key, fecha_hora)
		 VALUES ($1, $2, $3, 1, $4, $5)`,
		c.Tiquetera.ID, c.Cliente.ID, f.usuarioID, uuid.New(), at)
	require.NoError(t, err)
}

func ids(items []domain.ClienteReporte) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(items))
	for _, it := range items {
		out = append(out, it.ClienteID)
	}
	return out
}

func utc(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestReporteSinRecarga_BoundaryInBogota(t *testing.T) {
	f := newReporteFixture(t)
	day30 := f.client(t, 5, clienteTestDate, utc("2026-10-05T05:00:00Z"))
	day30Late := f.client(t, 5, clienteTestDate, utc("2026-10-06T04:59:59Z"))
	f.client(t, 5, clienteTestDate, utc("2026-10-06T05:00:00Z"))
	inactive := f.client(t, 5, clienteTestDate, utc("2026-09-01T05:00:00Z"))
	_, err := f.owner.Exec(t.Context(), `UPDATE clientes SET activo = false WHERE id = $1`, inactive.Cliente.ID)
	require.NoError(t, err)

	items, total, err := f.reportes.SinRecarga(t.Context(), 30, 1, 50)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, []uuid.UUID{day30.Cliente.ID, day30Late.Cliente.ID}, ids(items))
	assert.Equal(t, utc("2026-10-05T05:00:00Z"), items[0].Fecha.UTC())
}

func TestReportePorVencer_IncludesTodayAndLimitWithBalance(t *testing.T) {
	f := newReporteFixture(t)
	now := utc("2026-11-01T12:00:00Z")
	hoy := f.client(t, 2, "2026-11-04", now)
	limite := f.client(t, 2, "2026-11-11", now)
	f.client(t, 2, "2026-11-12", now)
	f.client(t, 2, "2026-11-03", now)
	f.client(t, 0, "2026-11-05", now)

	items, total, err := f.reportes.PorVencer(t.Context(), 7, 1, 50)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, []uuid.UUID{hoy.Cliente.ID, limite.Cliente.ID}, ids(items))
}

func TestReporteSaldoBajo_OnlyCurrentAtOrBelowMax(t *testing.T) {
	f := newReporteFixture(t)
	now := utc("2026-11-01T12:00:00Z")
	tres := f.client(t, 3, clienteTestDate, now)
	cero := f.client(t, 0, clienteTestDate, now)
	f.client(t, 4, clienteTestDate, now)
	f.client(t, 1, "2026-11-03", now)

	items, total, err := f.reportes.SaldoBajo(t.Context(), 3, 1, 50)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, []uuid.UUID{cero.Cliente.ID, tres.Cliente.ID}, ids(items))
}

func TestReporteDashboard_CountsTodayInBogota(t *testing.T) {
	f := newReporteFixture(t)
	c := f.client(t, 10, "2026-11-08", utc("2026-11-01T05:00:00Z"))
	f.client(t, 4, clienteTestDate, utc("2026-10-31T04:59:59Z"))
	f.client(t, 7, "2026-11-03", utc("2026-11-02T12:00:00Z"))
	f.consumoAt(t, c, utc("2026-11-05T04:30:00Z"))
	f.consumoAt(t, c, utc("2026-11-04T04:59:00Z"))
	f.consumoAt(t, c, utc("2026-10-29T05:00:00Z"))
	f.consumoAt(t, c, utc("2026-10-29T04:59:00Z"))
	_, err := f.owner.Exec(t.Context(),
		`INSERT INTO consumos_rechazados (cajero_id, motivo, fecha_hora) VALUES ($1, 'SIN_SALDO', '2026-11-04T06:00:00Z'), ($1, 'SIN_SALDO', '2026-11-04T04:00:00Z')`,
		f.usuarioID)
	require.NoError(t, err)

	d, err := f.reportes.Dashboard(t.Context())

	require.NoError(t, err)
	assert.Equal(t, int64(1), d.ConsumosHoy)
	require.Len(t, d.Consumos7d, 7)
	assert.Equal(t, "2026-10-29", d.Consumos7d[0].Fecha.Format(time.DateOnly))
	assert.Equal(t, int64(1), d.Consumos7d[0].Total)
	assert.Equal(t, int64(1), d.Consumos7d[5].Total)
	assert.Equal(t, int64(1), d.RechazosHoy)
	assert.Equal(t, domain.ResumenVigentes{Clientes: 2, SaldoPendiente: 14, PorVencer: 1}, d.Vigentes)
	assert.Equal(t, int64(40), d.VendidosMes)
}

func TestReporteVendidos_RechargeCountsOnlyNewLunches(t *testing.T) {
	f := newReporteFixture(t)
	c := f.client(t, 5, clienteTestDate, utc("2026-10-15T12:00:00Z"))
	_, err := f.svc.Recharge(t.Context(), f.usuarioID, c.Cliente.ID, domain.TiqueteraInput{Almuerzos: 10, FechaVencimiento: clienteTestDate})
	require.NoError(t, err)
	_, err = f.owner.Exec(t.Context(), `UPDATE tiqueteras SET created_at = '2026-11-02T12:00:00Z' WHERE cliente_id = $1 AND activa`, c.Cliente.ID)
	require.NoError(t, err)

	d, err := f.reportes.Dashboard(t.Context())

	require.NoError(t, err)
	assert.Equal(t, int64(10), d.VendidosMes)
	assert.Equal(t, int64(15), d.Vigentes.SaldoPendiente)
}

func TestReporteConsumos_GroupsByDayAndCashier(t *testing.T) {
	f := newReporteFixture(t)
	c := f.client(t, 10, clienteTestDate, utc("2026-11-01T05:00:00Z"))
	f.consumoAt(t, c, utc("2026-11-02T04:59:00Z"))
	f.consumoAt(t, c, utc("2026-11-02T05:00:00Z"))
	f.consumoAt(t, c, utc("2026-11-03T05:00:00Z"))

	r, err := f.reportes.Consumos(t.Context(), utc("2026-11-01T00:00:00Z"), utc("2026-11-02T00:00:00Z"), nil)

	require.NoError(t, err)
	assert.Equal(t, int64(2), r.Total)
	require.Len(t, r.Dias, 2)
	assert.Equal(t, int64(1), r.Dias[0].Total)
	assert.Equal(t, int64(1), r.Dias[1].Total)
	require.Len(t, r.Cajeros, 1)
	assert.Equal(t, f.usuarioID, r.Cajeros[0].CajeroID)

	other := uuid.New()
	r, err = f.reportes.Consumos(t.Context(), utc("2026-11-01T00:00:00Z"), utc("2026-11-02T00:00:00Z"), &other)
	require.NoError(t, err)
	assert.Equal(t, int64(0), r.Total)
	assert.Empty(t, r.Cajeros)
}
