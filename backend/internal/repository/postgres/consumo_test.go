package postgres

import (
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tiquetera/internal/domain"
	"tiquetera/internal/service"
)

const consumoTestPhone = "3001112233"

type consumoFixture struct {
	clienteFixture

	consumos *service.ConsumoService
	created  service.ClienteCreated
}

// newConsumoFixture has a client with the given balance. Its clock reads 22:00
// on 2026-11-04 in Bogotá, already the 5th in UTC.
func newConsumoFixture(t *testing.T, saldo int) consumoFixture {
	t.Helper()
	f := newClienteFixture(t)
	created, err := f.svc.Create(t.Context(), f.usuarioID,
		domain.ClienteData{Nombre: "Rosa Díaz", Telefono: consumoTestPhone},
		domain.TiqueteraInput{Almuerzos: saldo, FechaVencimiento: clienteTestDate})
	require.NoError(t, err)
	return consumoFixture{
		clienteFixture: f,
		consumos:       service.NewConsumoService(NewConsumoRepository(f.pool), f.clock),
		created:        created,
	}
}

func (f consumoFixture) consume(t *testing.T) (domain.ConsumoResult, error) {
	t.Helper()
	return f.consumos.Consume(t.Context(), f.usuarioID, uuid.New(), f.created.QRPayload)
}

func (f consumoFixture) balances(t *testing.T) (saldo, consumos int) {
	t.Helper()
	require.NoError(t, f.owner.QueryRow(t.Context(),
		`SELECT saldo FROM tiqueteras WHERE cliente_id = $1 AND activa`, f.created.Cliente.ID).Scan(&saldo))
	require.NoError(t, f.owner.QueryRow(t.Context(),
		`SELECT count(*) FROM consumos WHERE cliente_id = $1`, f.created.Cliente.ID).Scan(&consumos))
	return saldo, consumos
}

func (f consumoFixture) rejections(t *testing.T, code domain.RejectionCode) int {
	t.Helper()
	var n int
	require.NoError(t, f.owner.QueryRow(t.Context(),
		`SELECT count(*) FROM consumos_rechazados WHERE motivo = $1`, string(code)).Scan(&n))
	return n
}

func (f consumoFixture) setExpiry(t *testing.T, date string) {
	t.Helper()
	_, err := f.owner.Exec(t.Context(), `UPDATE tiqueteras SET fecha_vencimiento = $1::date WHERE cliente_id = $2`, date, f.created.Cliente.ID)
	require.NoError(t, err)
}

func assertRejected(t *testing.T, f consumoFixture, err error, code domain.RejectionCode) {
	t.Helper()
	var rejection domain.ConsumoRejection
	require.ErrorAs(t, err, &rejection)
	assert.Equal(t, code, rejection.Code)
	assert.Equal(t, 1, f.rejections(t, code))
}

func TestConsumo_Success_DebitsAndAppends(t *testing.T) {
	f := newConsumoFixture(t, 5)

	got, err := f.consume(t)
	require.NoError(t, err)
	assert.Equal(t, "Rosa Díaz", got.ClienteNombre)
	assert.Equal(t, int32(4), got.SaldoRestante)
	assert.Equal(t, clienteTestDate, got.FechaVencimiento.Format(time.DateOnly))
	assert.False(t, got.FechaHora.IsZero())

	saldo, consumos := f.balances(t)
	assert.Equal(t, 4, saldo)
	assert.Equal(t, 1, consumos)
}

func TestConsumo_TwentyConcurrentOnFive_ExactlyFiveSucceed(t *testing.T) {
	f := newConsumoFixture(t, 5)

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		remaining []int32
		noBalance int
		others    []error
	)
	for range 20 {
		wg.Go(func() {
			got, err := f.consume(t)
			mu.Lock()
			defer mu.Unlock()
			var rejection domain.ConsumoRejection
			switch {
			case err == nil:
				remaining = append(remaining, got.SaldoRestante)
			case errors.As(err, &rejection) && rejection.Code == domain.RejectNoBalance:
				noBalance++
			default:
				others = append(others, err)
			}
		})
	}
	wg.Wait()

	require.Empty(t, others)
	slices.Sort(remaining)
	assert.Equal(t, []int32{0, 1, 2, 3, 4}, remaining)
	assert.Equal(t, 15, noBalance)
	saldo, consumos := f.balances(t)
	assert.Zero(t, saldo)
	assert.Equal(t, 5, consumos)
	assert.Equal(t, 15, f.rejections(t, domain.RejectNoBalance))
}

func TestConsumo_ExpiryBoundaryInBogota(t *testing.T) {
	f := newConsumoFixture(t, 5)

	f.setExpiry(t, "2026-11-04")
	got, err := f.consume(t)
	require.NoError(t, err, "a book expiring today in Bogotá is accepted, although it is already the 5th in UTC")
	assert.Equal(t, int32(4), got.SaldoRestante)

	f.setExpiry(t, "2026-11-03")
	_, err = f.consume(t)
	assertRejected(t, f, err, domain.RejectTiqueteraExpired)
	var rejection domain.ConsumoRejection
	require.ErrorAs(t, err, &rejection)
	assert.Equal(t, "Tiquetera vencida el 03/11/2026", rejection.Message)

	saldo, _ := f.balances(t)
	assert.Equal(t, 4, saldo)
}

func TestConsumo_RepeatedKey_DoesNotDebitTwice(t *testing.T) {
	f := newConsumoFixture(t, 5)
	key := uuid.New()

	first, err := f.consumos.Consume(t.Context(), f.usuarioID, key, f.created.QRPayload)
	require.NoError(t, err)
	second, err := f.consumos.Consume(t.Context(), f.usuarioID, key, f.created.QRPayload)
	require.NoError(t, err)

	assert.Equal(t, first.SaldoRestante, second.SaldoRestante)
	assert.Equal(t, first.ClienteNombre, second.ClienteNombre)
	assert.True(t, first.FechaHora.Equal(second.FechaHora))
	saldo, consumos := f.balances(t)
	assert.Equal(t, 4, saldo)
	assert.Equal(t, 1, consumos)
}

func TestConsumo_ConcurrentSameKey_DebitsOnceAndAllGetSameResult(t *testing.T) {
	f := newConsumoFixture(t, 20)
	key := uuid.New()

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []int32
		errs    []error
	)
	for range 10 {
		wg.Go(func() {
			got, err := f.consumos.Consume(t.Context(), f.usuarioID, key, f.created.QRPayload)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				return
			}
			results = append(results, got.SaldoRestante)
		})
	}
	wg.Wait()

	require.Empty(t, errs)
	assert.Len(t, results, 10)
	for _, r := range results {
		assert.Equal(t, int32(19), r)
	}
	saldo, consumos := f.balances(t)
	assert.Equal(t, 19, saldo)
	assert.Equal(t, 1, consumos)
}

func TestConsumo_KeyOfAnotherUser_ReturnsConflict(t *testing.T) {
	f := newConsumoFixture(t, 5)
	key := uuid.New()
	_, err := f.consumos.Consume(t.Context(), f.usuarioID, key, f.created.QRPayload)
	require.NoError(t, err)

	var other uuid.UUID
	require.NoError(t, f.owner.QueryRow(t.Context(),
		`INSERT INTO usuarios (username, password_hash, nombre, rol) VALUES ('otro', 'x', 'Otro', 'cajero') RETURNING id`).Scan(&other))

	_, err = f.consumos.Consume(t.Context(), other, key, f.created.QRPayload)
	require.ErrorIs(t, err, domain.ErrIdempotencyConflict)
	saldo, _ := f.balances(t)
	assert.Equal(t, 4, saldo)
}

func TestConsumo_RechargeDuringDebits_LosesNothingAndDoesNotDeadlock(t *testing.T) {
	f := newConsumoFixture(t, 20)

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)
	record := func(err error) {
		if err != nil {
			mu.Lock()
			defer mu.Unlock()
			errs = append(errs, err)
		}
	}
	for i := range 10 {
		wg.Go(func() {
			_, err := f.consume(t)
			record(err)
		})
		if i == 3 {
			wg.Go(func() {
				_, err := f.svc.Recharge(t.Context(), f.usuarioID, f.created.Cliente.ID,
					domain.TiqueteraInput{Almuerzos: 5, FechaVencimiento: clienteTestDate})
				record(err)
			})
		}
	}
	wg.Wait()

	require.Empty(t, errs)
	saldo, consumos := f.balances(t)
	assert.Equal(t, 10, consumos)
	assert.Equal(t, 15, saldo, "20 + 5 recharged - 10 consumed")
}

func TestConsumo_Rejections_AreRecorded(t *testing.T) {
	t.Run("malformed QR", func(t *testing.T) {
		f := newConsumoFixture(t, 1)
		_, err := f.consumos.Consume(t.Context(), f.usuarioID, uuid.New(), "basura")
		assertRejected(t, f, err, domain.RejectQRInvalid)
	})
	t.Run("replaced QR", func(t *testing.T) {
		f := newConsumoFixture(t, 1)
		_, err := f.svc.RegenerateQR(t.Context(), f.usuarioID, f.created.Cliente.ID)
		require.NoError(t, err)
		_, err = f.consume(t)
		assertRejected(t, f, err, domain.RejectClienteNotFound)
		saldo, _ := f.balances(t)
		assert.Equal(t, 1, saldo)
	})
	t.Run("inactive client", func(t *testing.T) {
		f := newConsumoFixture(t, 1)
		inactive := false
		_, err := f.svc.Update(t.Context(), f.usuarioID, f.created.Cliente.ID, domain.ClientePatch{Activo: &inactive})
		require.NoError(t, err)
		_, err = f.consume(t)
		assertRejected(t, f, err, domain.RejectClienteInactive)
		saldo, _ := f.balances(t)
		assert.Equal(t, 1, saldo)
	})
	t.Run("no active book", func(t *testing.T) {
		f := newConsumoFixture(t, 1)
		_, err := f.owner.Exec(t.Context(), `UPDATE tiqueteras SET activa = false WHERE cliente_id = $1`, f.created.Cliente.ID)
		require.NoError(t, err)
		_, err = f.consume(t)
		assertRejected(t, f, err, domain.RejectNoTiquetera)
	})
	t.Run("no balance", func(t *testing.T) {
		f := newConsumoFixture(t, 1)
		_, err := f.consume(t)
		require.NoError(t, err)
		_, err = f.consume(t)
		assertRejected(t, f, err, domain.RejectNoBalance)
	})
}

func TestConsumo_Latency_ReportsP95(t *testing.T) {
	f := newConsumoFixture(t, 300)
	samples := make([]time.Duration, 0, 300)
	for range 300 {
		start := time.Now()
		_, err := f.consume(t)
		samples = append(samples, time.Since(start))
		require.NoError(t, err)
	}
	slices.Sort(samples)
	t.Logf("consumo p50=%s p95=%s max=%s", samples[150], samples[285], samples[299])
}
