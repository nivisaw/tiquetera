package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tiquetera/internal/domain"
)

type fakeConsumoRepo struct {
	replayResults []domain.ConsumoResult
	replayErr     error
	clienteErr    error
	cliente       domain.Cliente
	active        *domain.Tiquetera
	activeErr     error
	debitErrs     []error
	consumoErr    error
	rejectErr     error
	getErr        error
	rejections    []domain.RejectionCode
	rejectCliente []*uuid.UUID
	consumos      []domain.Consumo
	debitToday    time.Time
}

func (f *fakeConsumoRepo) InTx(_ context.Context, fn func(ConsumoRepository) error) error {
	return fn(f)
}

func (f *fakeConsumoRepo) GetConsumoResultado(context.Context, uuid.UUID) (domain.ConsumoResult, error) {
	if len(f.replayResults) > 0 {
		r := f.replayResults[0]
		f.replayResults = f.replayResults[1:]
		return r, nil
	}
	if f.replayErr != nil {
		return domain.ConsumoResult{}, f.replayErr
	}
	return domain.ConsumoResult{}, domain.ErrNotFound
}

func (f *fakeConsumoRepo) GetClienteByQRHash(context.Context, []byte) (domain.Cliente, error) {
	return f.cliente, f.clienteErr
}

func (f *fakeConsumoRepo) GetCliente(context.Context, uuid.UUID) (domain.Cliente, error) {
	return f.cliente, f.getErr
}

func (f *fakeConsumoRepo) GetActiveTiquetera(context.Context, uuid.UUID) (domain.Tiquetera, error) {
	if f.activeErr != nil {
		return domain.Tiquetera{}, f.activeErr
	}
	if f.active == nil {
		return domain.Tiquetera{}, domain.ErrNotFound
	}
	return *f.active, nil
}

func (f *fakeConsumoRepo) DebitTiquetera(_ context.Context, _ uuid.UUID, today time.Time) (domain.Tiquetera, error) {
	f.debitToday = today
	if len(f.debitErrs) > 0 {
		err := f.debitErrs[0]
		f.debitErrs = f.debitErrs[1:]
		if err != nil {
			return domain.Tiquetera{}, err
		}
	}
	return domain.Tiquetera{ID: uuid.New(), Saldo: 4, FechaVencimiento: consumoExpiry}, nil
}

func (f *fakeConsumoRepo) CreateConsumo(_ context.Context, c domain.Consumo) (time.Time, error) {
	f.consumos = append(f.consumos, c)
	return consumoTime, f.consumoErr
}

func (f *fakeConsumoRepo) AddRejection(_ context.Context, _ uuid.UUID, cliente *uuid.UUID, code domain.RejectionCode) error {
	f.rejections = append(f.rejections, code)
	f.rejectCliente = append(f.rejectCliente, cliente)
	return f.rejectErr
}

var (
	consumoExpiry = time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	consumoTime   = time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
)

func newConsumoSvc(repo *fakeConsumoRepo) *ConsumoService {
	// 22:00 on Nov 4 in Bogotá, already Nov 5 in UTC.
	return NewConsumoService(repo, &fakeClock{t: time.Date(2026, 11, 5, 3, 0, 0, 0, time.UTC)})
}

func validQR() string {
	payload, _ := domain.NewQRToken()
	return payload
}

func rejectionCode(t *testing.T, err error) domain.RejectionCode {
	t.Helper()
	var rejection domain.ConsumoRejection
	require.ErrorAs(t, err, &rejection)
	return rejection.Code
}

func TestConsume_Eligible_DebitsAndStoresConsumo(t *testing.T) {
	repo := &fakeConsumoRepo{cliente: domain.Cliente{ClienteData: domain.ClienteData{Nombre: "Ana"}, ID: uuid.New(), Activo: true}}
	cajero, key := uuid.New(), uuid.New()

	got, err := newConsumoSvc(repo).Consume(t.Context(), cajero, key, validQR())
	require.NoError(t, err)

	assert.Equal(t, "Ana", got.ClienteNombre)
	assert.Equal(t, int32(4), got.SaldoRestante)
	assert.Equal(t, consumoTime, got.FechaHora)
	assert.Equal(t, time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC), repo.debitToday, "today must be the Bogotá date")
	require.Len(t, repo.consumos, 1)
	assert.Equal(t, key, repo.consumos[0].IdempotencyKey)
	assert.Equal(t, cajero, repo.consumos[0].CajeroID)
	assert.Empty(t, repo.rejections)
}

func TestConsume_RepeatedKey_ReturnsOriginalWithoutDebit(t *testing.T) {
	cajero := uuid.New()
	original := domain.ConsumoResult{CajeroID: cajero, SaldoRestante: 9}
	repo := &fakeConsumoRepo{replayResults: []domain.ConsumoResult{original}}

	got, err := newConsumoSvc(repo).Consume(t.Context(), cajero, uuid.New(), "cualquier cosa")
	require.NoError(t, err)
	assert.Equal(t, original, got)
	assert.Empty(t, repo.consumos)
	assert.Empty(t, repo.rejections)
}

func TestConsume_KeyOfAnotherUser_ReturnsConflict(t *testing.T) {
	repo := &fakeConsumoRepo{replayResults: []domain.ConsumoResult{{CajeroID: uuid.New(), ClienteNombre: "Ana"}}}
	got, err := newConsumoSvc(repo).Consume(t.Context(), uuid.New(), uuid.New(), validQR())
	require.ErrorIs(t, err, domain.ErrIdempotencyConflict)
	assert.Empty(t, got.ClienteNombre)
}

func TestConsume_ConcurrentDuplicateKey_ReturnsOriginal(t *testing.T) {
	cajero := uuid.New()
	winner := domain.ConsumoResult{CajeroID: cajero, SaldoRestante: 2}
	repo := &duplicateKeyRepo{fakeConsumoRepo: fakeConsumoRepo{consumoErr: domain.ErrDuplicateIdempotencyKey}, winner: winner}

	got, err := NewConsumoService(repo, &fakeClock{t: consumoTime}).Consume(t.Context(), cajero, uuid.New(), validQR())
	require.NoError(t, err)
	assert.Equal(t, winner, got)
}

// duplicateKeyRepo misses the first lookup and finds the winner after the conflict.
type duplicateKeyRepo struct {
	fakeConsumoRepo

	winner  domain.ConsumoResult
	lookups int
}

func (r *duplicateKeyRepo) InTx(_ context.Context, fn func(ConsumoRepository) error) error {
	return fn(r)
}

func (r *duplicateKeyRepo) GetConsumoResultado(context.Context, uuid.UUID) (domain.ConsumoResult, error) {
	r.lookups++
	if r.lookups == 1 {
		return domain.ConsumoResult{}, domain.ErrNotFound
	}
	return r.winner, nil
}

func TestConsume_MalformedQR_RecordsRejectionWithoutClient(t *testing.T) {
	repo := &fakeConsumoRepo{}
	_, err := newConsumoSvc(repo).Consume(t.Context(), uuid.New(), uuid.New(), "basura")
	assert.Equal(t, domain.RejectQRInvalid, rejectionCode(t, err))
	assert.Equal(t, []domain.RejectionCode{domain.RejectQRInvalid}, repo.rejections)
	assert.Nil(t, repo.rejectCliente[0])
}

func TestConsume_UnknownQR_RecordsRejectionWithoutClient(t *testing.T) {
	repo := &fakeConsumoRepo{clienteErr: domain.ErrNotFound}
	_, err := newConsumoSvc(repo).Consume(t.Context(), uuid.New(), uuid.New(), validQR())
	assert.Equal(t, domain.RejectClienteNotFound, rejectionCode(t, err))
	assert.Nil(t, repo.rejectCliente[0])
}

func TestConsume_NotEligible_RecordsReasonWithClient(t *testing.T) {
	today := time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		activo bool
		book   *domain.Tiquetera
		code   domain.RejectionCode
	}{
		{"inactive", false, &domain.Tiquetera{Saldo: 1, FechaVencimiento: today}, domain.RejectClienteInactive},
		{"no book", true, nil, domain.RejectNoTiquetera},
		{"expired yesterday", true, &domain.Tiquetera{Saldo: 1, FechaVencimiento: today.AddDate(0, 0, -1)}, domain.RejectTiqueteraExpired},
		{"no balance", true, &domain.Tiquetera{Saldo: 0, FechaVencimiento: today}, domain.RejectNoBalance},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := uuid.New()
			repo := &fakeConsumoRepo{
				cliente:   domain.Cliente{ID: id, Activo: tt.activo},
				active:    tt.book,
				debitErrs: []error{domain.ErrNotFound},
			}
			_, err := newConsumoSvc(repo).Consume(t.Context(), uuid.New(), uuid.New(), validQR())
			assert.Equal(t, tt.code, rejectionCode(t, err))
			assert.Equal(t, []domain.RejectionCode{tt.code}, repo.rejections)
			assert.Equal(t, id, *repo.rejectCliente[0])
			assert.Empty(t, repo.consumos)
		})
	}
}

func TestConsume_BookReplacedDuringDebit_RetriesAndDebits(t *testing.T) {
	today := time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
	repo := &fakeConsumoRepo{
		cliente:   domain.Cliente{Activo: true},
		active:    &domain.Tiquetera{Saldo: 5, FechaVencimiento: today},
		debitErrs: []error{domain.ErrNotFound, nil},
	}
	got, err := newConsumoSvc(repo).Consume(t.Context(), uuid.New(), uuid.New(), validQR())
	require.NoError(t, err)
	assert.Equal(t, int32(4), got.SaldoRestante)
	assert.Empty(t, repo.rejections)
	assert.Len(t, repo.consumos, 1)
}

func TestConsume_BookKeepsChanging_FailsWithoutRejection(t *testing.T) {
	today := time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
	repo := &fakeConsumoRepo{
		cliente:   domain.Cliente{Activo: true},
		active:    &domain.Tiquetera{Saldo: 5, FechaVencimiento: today},
		debitErrs: []error{domain.ErrNotFound, domain.ErrNotFound, domain.ErrNotFound},
	}
	_, err := newConsumoSvc(repo).Consume(t.Context(), uuid.New(), uuid.New(), validQR())
	require.ErrorIs(t, err, errDebitRaced)
	assert.Empty(t, repo.rejections)
}

func TestConsume_StorageFailures_AreReturned(t *testing.T) {
	notEligible := []error{domain.ErrNotFound}
	tests := []struct {
		name string
		repo *fakeConsumoRepo
	}{
		{"replay", &fakeConsumoRepo{replayErr: errRepo}},
		{"client lookup", &fakeConsumoRepo{clienteErr: errRepo}},
		{"debit", &fakeConsumoRepo{debitErrs: []error{errRepo}}},
		{"insert", &fakeConsumoRepo{consumoErr: errRepo}},
		{"diagnose client", &fakeConsumoRepo{debitErrs: notEligible, getErr: errRepo}},
		{"diagnose book", &fakeConsumoRepo{cliente: domain.Cliente{Activo: true}, debitErrs: notEligible, activeErr: errRepo}},
		{"record rejection", &fakeConsumoRepo{debitErrs: notEligible, rejectErr: errRepo}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newConsumoSvc(tt.repo).Consume(t.Context(), uuid.New(), uuid.New(), validQR())
			require.ErrorIs(t, err, errRepo)
		})
	}
}
