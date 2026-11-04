package postgres

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tiquetera/internal/domain"
	"tiquetera/internal/service"
)

const (
	clienteTestDate  = "2026-12-31"
	clienteTestPhone = "3001112233"
)

type clienteFixture struct {
	fixture

	svc *service.ClienteService
}

func newClienteFixture(t *testing.T) clienteFixture {
	t.Helper()
	f := setup(t)
	pool, err := pgxpool.New(t.Context(), f.app.Config().ConnString())
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	clock := &movableClock{t: time.Date(2026, 11, 5, 3, 0, 0, 0, time.UTC)}
	return clienteFixture{fixture: f, svc: service.NewClienteService(NewClienteRepository(pool), clock)}
}

func (f clienteFixture) create(t *testing.T, phone string) service.ClienteCreated {
	t.Helper()
	created, err := f.svc.Create(t.Context(), f.usuarioID,
		domain.ClienteData{Nombre: "Rosa Díaz", Telefono: phone},
		domain.TiqueteraInput{Almuerzos: 20, FechaVencimiento: clienteTestDate})
	require.NoError(t, err)
	return created
}

func (f clienteFixture) storedHash(t *testing.T, id uuid.UUID) []byte {
	t.Helper()
	var hash []byte
	require.NoError(t, f.owner.QueryRow(t.Context(), `SELECT qr_token_hash FROM clientes WHERE id = $1`, id).Scan(&hash))
	return hash
}

func TestClienteCreate_StoresOnlyTokenHashAndAudits(t *testing.T) {
	f := newClienteFixture(t)
	created := f.create(t, clienteTestPhone)

	assert.Equal(t, "+573001112233", created.Cliente.Telefono)
	assert.Equal(t, int32(20), created.Tiquetera.Saldo)
	assert.True(t, created.Tiquetera.Activa)
	assert.Equal(t, clienteTestDate, created.Tiquetera.FechaVencimiento.Format(time.DateOnly))

	token := strings.TrimPrefix(created.QRPayload, "TQ1.")
	assert.Equal(t, domain.HashQRToken(token), f.storedHash(t, created.Cliente.ID))

	var audited int
	require.NoError(t, f.owner.QueryRow(t.Context(),
		`SELECT count(*) FROM audit_log WHERE accion = $1 AND entidad_id = $2 AND datos = '{}'`,
		domain.AuditClienteCreated, created.Cliente.ID).Scan(&audited))
	assert.Equal(t, 1, audited)
}

func TestClienteCreate_DuplicatePhone_ReturnsErrPhoneTakenAndWritesNothing(t *testing.T) {
	f := newClienteFixture(t)
	f.create(t, clienteTestPhone)

	_, err := f.svc.Create(t.Context(), f.usuarioID,
		domain.ClienteData{Nombre: "Otra", Telefono: "+57 300 111 2233"},
		domain.TiqueteraInput{Almuerzos: 1, FechaVencimiento: clienteTestDate})
	require.ErrorIs(t, err, domain.ErrPhoneTaken)

	var count int
	require.NoError(t, f.owner.QueryRow(t.Context(), `SELECT count(*) FROM clientes WHERE nombre = 'Otra'`).Scan(&count))
	assert.Zero(t, count)
}

func TestClienteCreate_ExpiryBoundaryInBogota(t *testing.T) {
	f := newClienteFixture(t)
	input := func(date string) domain.TiqueteraInput {
		return domain.TiqueteraInput{Almuerzos: 1, FechaVencimiento: date}
	}
	_, err := f.svc.Create(t.Context(), f.usuarioID, domain.ClienteData{Nombre: "Hoy", Telefono: "3001112244"}, input("2026-11-04"))
	require.NoError(t, err)
	_, err = f.svc.Create(t.Context(), f.usuarioID, domain.ClienteData{Nombre: "Ayer", Telefono: "3001112255"}, input("2026-11-03"))
	assert.ErrorAs(t, err, &domain.ValidationError{})
}

func TestClienteRecharge_ReplacesActiveTiquetera(t *testing.T) {
	f := newClienteFixture(t)
	created := f.create(t, clienteTestPhone)

	recharged, err := f.svc.Recharge(t.Context(), f.usuarioID, created.Cliente.ID,
		domain.TiqueteraInput{Almuerzos: 5, FechaVencimiento: clienteTestDate})
	require.NoError(t, err)

	detail, err := f.svc.Get(t.Context(), created.Cliente.ID)
	require.NoError(t, err)
	require.NotNil(t, detail.Tiquetera)
	assert.Equal(t, recharged.ID, detail.Tiquetera.ID)
	assert.Equal(t, int32(25), detail.Tiquetera.Saldo)
	assert.Equal(t, int32(25), detail.Tiquetera.AlmuerzosComprados)

	_, err = f.svc.Recharge(t.Context(), f.usuarioID, created.Cliente.ID,
		domain.TiqueteraInput{Almuerzos: 476, FechaVencimiento: clienteTestDate})
	require.ErrorAs(t, err, &domain.ValidationError{})
	detail, err = f.svc.Get(t.Context(), created.Cliente.ID)
	require.NoError(t, err)
	require.NotNil(t, detail.Tiquetera, "a rejected recharge must roll back the deactivation")
	assert.Equal(t, recharged.ID, detail.Tiquetera.ID)

	var active int
	require.NoError(t, f.owner.QueryRow(t.Context(),
		`SELECT count(*) FROM tiqueteras WHERE cliente_id = $1 AND activa`, created.Cliente.ID).Scan(&active))
	assert.Equal(t, 1, active)
}

func TestClienteRecharge_InactiveCliente_KeepsTiquetera(t *testing.T) {
	f := newClienteFixture(t)
	created := f.create(t, clienteTestPhone)
	inactive := false
	_, err := f.svc.Update(t.Context(), f.usuarioID, created.Cliente.ID, domain.ClientePatch{Activo: &inactive})
	require.NoError(t, err)

	_, err = f.svc.Recharge(t.Context(), f.usuarioID, created.Cliente.ID,
		domain.TiqueteraInput{Almuerzos: 5, FechaVencimiento: clienteTestDate})
	require.ErrorIs(t, err, domain.ErrClienteInactive)

	detail, err := f.svc.Get(t.Context(), created.Cliente.ID)
	require.NoError(t, err)
	require.NotNil(t, detail.Tiquetera)
	assert.Equal(t, created.Tiquetera.ID, detail.Tiquetera.ID)
}

func TestClienteRegenerateQR_InvalidatesPreviousHash(t *testing.T) {
	f := newClienteFixture(t)
	created := f.create(t, clienteTestPhone)
	before := f.storedHash(t, created.Cliente.ID)

	payload, err := f.svc.RegenerateQR(t.Context(), f.usuarioID, created.Cliente.ID)
	require.NoError(t, err)

	after := f.storedHash(t, created.Cliente.ID)
	assert.NotEqual(t, before, after)
	assert.Equal(t, domain.HashQRToken(strings.TrimPrefix(payload, "TQ1.")), after)

	detail, err := f.svc.Get(t.Context(), created.Cliente.ID)
	require.NoError(t, err)
	assert.Equal(t, int32(2), detail.Cliente.QRVersion)
}

func TestClienteRegenerateQR_Unknown_ReturnsNotFound(t *testing.T) {
	f := newClienteFixture(t)
	_, err := f.svc.RegenerateQR(t.Context(), f.usuarioID, uuid.New())
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestClienteUpdate_ClearsOptionalAndDetectsDuplicatePhone(t *testing.T) {
	f := newClienteFixture(t)
	email := "rosa@example.com"
	created, err := f.svc.Create(t.Context(), f.usuarioID,
		domain.ClienteData{Nombre: "Rosa", Telefono: clienteTestPhone, Email: &email},
		domain.TiqueteraInput{Almuerzos: 1, FechaVencimiento: clienteTestDate})
	require.NoError(t, err)

	updated, err := f.svc.Update(t.Context(), f.usuarioID, created.Cliente.ID, domain.ClientePatch{SetEmail: true})
	require.NoError(t, err)
	assert.Nil(t, updated.Email)
	assert.Equal(t, "Rosa", updated.Nombre)

	taken := "+573001234567"
	_, err = f.svc.Update(t.Context(), f.usuarioID, created.Cliente.ID, domain.ClientePatch{Telefono: &taken})
	assert.ErrorIs(t, err, domain.ErrPhoneTaken)
}

func TestClienteList_PaginatesAndTreatsWildcardsLiterally(t *testing.T) {
	f := newClienteFixture(t)
	f.create(t, clienteTestPhone)
	f.create(t, "3001112244")

	items, total, err := f.svc.List(t.Context(), "rosa", 2, 1)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, int64(2), total)

	_, total, err = f.svc.List(t.Context(), "300111", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)

	items, total, err = f.svc.List(t.Context(), "%", 1, 10)
	require.NoError(t, err)
	assert.Empty(t, items)
	assert.Zero(t, total)
}
