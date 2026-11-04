package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tiquetera/internal/domain"
)

var errRepo = errors.New("repo down")

// fakeClienteRepo records calls. failOn makes the named method fail with err.
type fakeClienteRepo struct {
	cliente       domain.Cliente
	tiquetera     domain.Tiquetera
	getErr        error
	tiqueteraErr  error
	previousSaldo int32
	failOn        string
	calls         []string
	hash          []byte
	audits        []domain.AuditEntry
	limit         int32
	offset        int32
}

func (f *fakeClienteRepo) step(name string) error {
	f.calls = append(f.calls, name)
	if f.failOn == name {
		return errRepo
	}
	return nil
}

func (f *fakeClienteRepo) InTx(_ context.Context, fn func(ClienteRepository) error) error {
	return fn(f)
}

func (f *fakeClienteRepo) CreateCliente(_ context.Context, data domain.ClienteData, hash []byte, _ uuid.UUID) (domain.Cliente, error) {
	f.hash = hash
	return domain.Cliente{ClienteData: data, ID: uuid.New(), Activo: true}, f.step("CreateCliente")
}

func (f *fakeClienteRepo) GetCliente(context.Context, uuid.UUID) (domain.Cliente, error) {
	return f.cliente, f.getErr
}

func (f *fakeClienteRepo) GetClienteForUpdate(context.Context, uuid.UUID) (domain.Cliente, error) {
	return f.cliente, f.getErr
}

func (f *fakeClienteRepo) SearchClientes(_ context.Context, _ string, limit, offset int32) ([]domain.Cliente, int64, error) {
	f.limit, f.offset = limit, offset
	return []domain.Cliente{f.cliente}, 7, f.step("SearchClientes")
}

func (f *fakeClienteRepo) UpdateCliente(_ context.Context, c domain.Cliente) (domain.Cliente, error) {
	return c, f.step("UpdateCliente")
}

func (f *fakeClienteRepo) RotateQR(_ context.Context, _ uuid.UUID, hash []byte) error {
	f.hash = hash
	return f.step("RotateQR")
}

func (f *fakeClienteRepo) GetActiveTiquetera(context.Context, uuid.UUID) (domain.Tiquetera, error) {
	return f.tiquetera, f.tiqueteraErr
}

func (f *fakeClienteRepo) DeactivateTiquetera(context.Context, uuid.UUID) (int32, error) {
	return f.previousSaldo, f.step("DeactivateTiquetera")
}

func (f *fakeClienteRepo) CreateTiquetera(_ context.Context, t domain.Tiquetera, _ uuid.UUID) (domain.Tiquetera, error) {
	t.ID = uuid.New()
	return t, f.step("CreateTiquetera")
}

func (f *fakeClienteRepo) AddAudit(_ context.Context, entry domain.AuditEntry) error {
	f.audits = append(f.audits, entry)
	return f.step("AddAudit")
}

func newClienteSvc(repo *fakeClienteRepo) *ClienteService {
	return NewClienteService(repo, &fakeClock{t: time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)})
}

var (
	validData  = domain.ClienteData{Nombre: "Ana", Telefono: "3001234567"}
	validInput = domain.TiqueteraInput{Almuerzos: 10, FechaVencimiento: "2026-12-01"}
)

func TestClienteServiceCreate_Valid_ReturnsPayloadMatchingStoredHash(t *testing.T) {
	repo := &fakeClienteRepo{}
	actor := uuid.New()
	got, err := newClienteSvc(repo).Create(t.Context(), actor, validData, validInput)
	require.NoError(t, err)

	assert.Equal(t, "+573001234567", got.Cliente.Telefono)
	assert.Equal(t, int32(10), got.Tiquetera.Saldo)
	assert.Equal(t, int32(10), got.Tiquetera.AlmuerzosComprados)
	token := strings.TrimPrefix(got.QRPayload, "TQ1.")
	assert.Equal(t, domain.HashQRToken(token), repo.hash)
	require.Len(t, repo.audits, 1)
	assert.Equal(t, domain.AuditClienteCreated, repo.audits[0].Accion)
	assert.Equal(t, actor, *repo.audits[0].UsuarioID)
}

func TestClienteServiceCreate_Invalid_DoesNotTouchRepo(t *testing.T) {
	tests := []struct {
		name  string
		data  domain.ClienteData
		input domain.TiqueteraInput
	}{
		{"bad phone", domain.ClienteData{Nombre: "Ana", Telefono: "1"}, validInput},
		{"too many lunches", validData, domain.TiqueteraInput{Almuerzos: 501, FechaVencimiento: "2026-12-01"}},
		{"expired", validData, domain.TiqueteraInput{Almuerzos: 1, FechaVencimiento: "2026-11-03"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeClienteRepo{}
			_, err := newClienteSvc(repo).Create(t.Context(), uuid.New(), tt.data, tt.input)
			assert.ErrorAs(t, err, &domain.ValidationError{})
			assert.Empty(t, repo.calls)
		})
	}
}

func TestClienteServiceCreate_StepFails_ReturnsErrorAndNoPayload(t *testing.T) {
	for _, step := range []string{"CreateCliente", "CreateTiquetera", "AddAudit"} {
		t.Run(step, func(t *testing.T) {
			got, err := newClienteSvc(&fakeClienteRepo{failOn: step}).Create(t.Context(), uuid.New(), validData, validInput)
			require.ErrorIs(t, err, errRepo)
			assert.Empty(t, got.QRPayload)
		})
	}
}

func TestClienteServiceGet(t *testing.T) {
	t.Run("with active book", func(t *testing.T) {
		got, err := newClienteSvc(&fakeClienteRepo{tiquetera: domain.Tiquetera{Saldo: 3}}).Get(t.Context(), uuid.New())
		require.NoError(t, err)
		require.NotNil(t, got.Tiquetera)
		assert.Equal(t, int32(3), got.Tiquetera.Saldo)
	})
	t.Run("without active book", func(t *testing.T) {
		got, err := newClienteSvc(&fakeClienteRepo{tiqueteraErr: domain.ErrNotFound}).Get(t.Context(), uuid.New())
		require.NoError(t, err)
		assert.Nil(t, got.Tiquetera)
	})
	t.Run("unknown client", func(t *testing.T) {
		_, err := newClienteSvc(&fakeClienteRepo{getErr: domain.ErrNotFound}).Get(t.Context(), uuid.New())
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
	t.Run("book lookup fails", func(t *testing.T) {
		_, err := newClienteSvc(&fakeClienteRepo{tiqueteraErr: errRepo}).Get(t.Context(), uuid.New())
		assert.ErrorIs(t, err, errRepo)
	})
}

func TestClienteServiceList_ComputesOffset(t *testing.T) {
	repo := &fakeClienteRepo{}
	items, total, err := newClienteSvc(repo).List(t.Context(), "a", 3, 20)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, int64(7), total)
	assert.Equal(t, int32(20), repo.limit)
	assert.Equal(t, int32(40), repo.offset)

	_, _, err = newClienteSvc(&fakeClienteRepo{failOn: "SearchClientes"}).List(t.Context(), "", 1, 20)
	assert.ErrorIs(t, err, errRepo)
}

func TestClienteServiceUpdate(t *testing.T) {
	existing := domain.Cliente{ClienteData: domain.ClienteData{Nombre: "Ana", Telefono: "+573001234567"}, Activo: true}

	t.Run("applies patch and audits", func(t *testing.T) {
		repo := &fakeClienteRepo{cliente: existing}
		nombre, inactive := " Rosa ", false
		got, err := newClienteSvc(repo).Update(t.Context(), uuid.New(), uuid.New(), domain.ClientePatch{Nombre: &nombre, Activo: &inactive})
		require.NoError(t, err)
		assert.Equal(t, "Rosa", got.Nombre)
		assert.False(t, got.Activo)
		assert.Equal(t, domain.AuditClienteUpdated, repo.audits[0].Accion)
	})
	t.Run("invalid value does not write", func(t *testing.T) {
		repo := &fakeClienteRepo{cliente: existing}
		bad := "1"
		_, err := newClienteSvc(repo).Update(t.Context(), uuid.New(), uuid.New(), domain.ClientePatch{Telefono: &bad})
		require.ErrorAs(t, err, &domain.ValidationError{})
		assert.Empty(t, repo.calls)
	})
	t.Run("unknown client", func(t *testing.T) {
		repo := &fakeClienteRepo{getErr: domain.ErrNotFound}
		_, err := newClienteSvc(repo).Update(t.Context(), uuid.New(), uuid.New(), domain.ClientePatch{})
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
	t.Run("write fails", func(t *testing.T) {
		repo := &fakeClienteRepo{cliente: existing, failOn: "UpdateCliente"}
		_, err := newClienteSvc(repo).Update(t.Context(), uuid.New(), uuid.New(), domain.ClientePatch{})
		assert.ErrorIs(t, err, errRepo)
	})
}

func TestClienteServiceRecharge(t *testing.T) {
	active := domain.Cliente{Activo: true}

	t.Run("deactivates before creating", func(t *testing.T) {
		repo := &fakeClienteRepo{cliente: active}
		id := uuid.New()
		got, err := newClienteSvc(repo).Recharge(t.Context(), uuid.New(), id, validInput)
		require.NoError(t, err)
		assert.Equal(t, id, got.ClienteID)
		assert.Equal(t, int32(10), got.Saldo)
		assert.Equal(t, []string{"DeactivateTiquetera", "CreateTiquetera", "AddAudit"}, repo.calls)
		assert.Equal(t, domain.AuditEntityTiquetera, repo.audits[0].Entidad)
	})
	t.Run("carries previous balance", func(t *testing.T) {
		repo := &fakeClienteRepo{cliente: active, previousSaldo: 3}
		got, err := newClienteSvc(repo).Recharge(t.Context(), uuid.New(), uuid.New(), validInput)
		require.NoError(t, err)
		assert.Equal(t, int32(13), got.Saldo)
		assert.Equal(t, int32(13), got.AlmuerzosComprados)
	})
	t.Run("exactly the limit is accepted", func(t *testing.T) {
		repo := &fakeClienteRepo{cliente: active, previousSaldo: 490}
		got, err := newClienteSvc(repo).Recharge(t.Context(), uuid.New(), uuid.New(), validInput)
		require.NoError(t, err)
		assert.Equal(t, int32(500), got.Saldo)
	})
	t.Run("over the limit is rejected", func(t *testing.T) {
		repo := &fakeClienteRepo{cliente: active, previousSaldo: 491}
		_, err := newClienteSvc(repo).Recharge(t.Context(), uuid.New(), uuid.New(), validInput)
		require.ErrorAs(t, err, &domain.ValidationError{})
		assert.Equal(t, []string{"DeactivateTiquetera"}, repo.calls)
	})
	t.Run("invalid input", func(t *testing.T) {
		repo := &fakeClienteRepo{cliente: active}
		_, err := newClienteSvc(repo).Recharge(t.Context(), uuid.New(), uuid.New(), domain.TiqueteraInput{})
		require.ErrorAs(t, err, &domain.ValidationError{})
		assert.Empty(t, repo.calls)
	})
	t.Run("inactive client", func(t *testing.T) {
		repo := &fakeClienteRepo{}
		_, err := newClienteSvc(repo).Recharge(t.Context(), uuid.New(), uuid.New(), validInput)
		require.ErrorIs(t, err, domain.ErrClienteInactive)
		assert.Empty(t, repo.calls)
	})
	t.Run("unknown client", func(t *testing.T) {
		repo := &fakeClienteRepo{getErr: domain.ErrNotFound}
		_, err := newClienteSvc(repo).Recharge(t.Context(), uuid.New(), uuid.New(), validInput)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})
	for _, step := range []string{"DeactivateTiquetera", "CreateTiquetera", "AddAudit"} {
		t.Run(step+" fails", func(t *testing.T) {
			repo := &fakeClienteRepo{cliente: active, failOn: step}
			_, err := newClienteSvc(repo).Recharge(t.Context(), uuid.New(), uuid.New(), validInput)
			assert.ErrorIs(t, err, errRepo)
		})
	}
}

func TestClienteServiceRegenerateQR(t *testing.T) {
	repo := &fakeClienteRepo{}
	payload, err := newClienteSvc(repo).RegenerateQR(t.Context(), uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, domain.HashQRToken(strings.TrimPrefix(payload, "TQ1.")), repo.hash)
	assert.Equal(t, domain.AuditQRRegenerated, repo.audits[0].Accion)

	got, err := newClienteSvc(&fakeClienteRepo{failOn: "RotateQR"}).RegenerateQR(t.Context(), uuid.New(), uuid.New())
	require.ErrorIs(t, err, errRepo)
	assert.Empty(t, got)
}
