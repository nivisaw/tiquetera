package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"tiquetera/internal/domain"
	"tiquetera/internal/platform"
)

// ClienteRepository is the persistence port for clients and ticket books.
type ClienteRepository interface {
	InTx(ctx context.Context, fn func(ClienteRepository) error) error
	CreateCliente(ctx context.Context, data domain.ClienteData, qrHash []byte, actor uuid.UUID) (domain.Cliente, error)
	GetCliente(ctx context.Context, id uuid.UUID) (domain.Cliente, error)
	GetClienteForUpdate(ctx context.Context, id uuid.UUID) (domain.Cliente, error)
	SearchClientes(ctx context.Context, q string, limit, offset int32) ([]domain.Cliente, int64, error)
	UpdateCliente(ctx context.Context, c domain.Cliente) (domain.Cliente, error)
	RotateQR(ctx context.Context, id uuid.UUID, qrHash []byte) error
	GetActiveTiquetera(ctx context.Context, clienteID uuid.UUID) (domain.Tiquetera, error)
	DeactivateTiquetera(ctx context.Context, clienteID uuid.UUID) (int32, error)
	CreateTiquetera(ctx context.Context, t domain.Tiquetera, actor uuid.UUID) (domain.Tiquetera, error)
	ListConsumos(ctx context.Context, clienteID uuid.UUID, limit, offset int32) ([]domain.ConsumoEntry, int64, error)
	AddAudit(ctx context.Context, entry domain.AuditEntry) error
}

// ClienteService implements client registration and ticket book management.
type ClienteService struct {
	repo  ClienteRepository
	clock platform.Clock
}

// NewClienteService returns a ClienteService.
func NewClienteService(repo ClienteRepository, clock platform.Clock) *ClienteService {
	return &ClienteService{repo: repo, clock: clock}
}

// ClienteCreated is the result of a registration. QRPayload is only available here.
type ClienteCreated struct {
	Cliente   domain.Cliente
	Tiquetera domain.Tiquetera
	QRPayload string
}

// ClienteDetail is a client with its active ticket book, if any.
type ClienteDetail struct {
	Cliente   domain.Cliente
	Tiquetera *domain.Tiquetera
}

// Create registers a client with its first ticket book atomically.
func (s *ClienteService) Create(ctx context.Context, actor uuid.UUID, data domain.ClienteData, input domain.TiqueteraInput) (ClienteCreated, error) {
	data, err := data.Normalize()
	if err != nil {
		return ClienteCreated{}, err
	}
	fecha, err := input.Validate(s.clock.Now())
	if err != nil {
		return ClienteCreated{}, err
	}
	payload, hash := domain.NewQRToken()

	result := ClienteCreated{QRPayload: payload}
	err = s.repo.InTx(ctx, func(tx ClienteRepository) error {
		var txErr error
		if result.Cliente, txErr = tx.CreateCliente(ctx, data, hash, actor); txErr != nil {
			return txErr
		}
		if result.Tiquetera, txErr = createTiquetera(ctx, tx, result.Cliente.ID, input.Almuerzos, fecha, actor); txErr != nil {
			return txErr
		}
		return auditCliente(ctx, tx, actor, domain.AuditClienteCreated, domain.AuditEntityCliente, result.Cliente.ID)
	})
	if err != nil {
		return ClienteCreated{}, fmt.Errorf("creating client: %w", err)
	}
	return result, nil
}

// Get returns a client with its active ticket book.
func (s *ClienteService) Get(ctx context.Context, id uuid.UUID) (ClienteDetail, error) {
	cliente, err := s.repo.GetCliente(ctx, id)
	if err != nil {
		return ClienteDetail{}, fmt.Errorf("getting client: %w", err)
	}
	detail := ClienteDetail{Cliente: cliente}
	tiquetera, err := s.repo.GetActiveTiquetera(ctx, id)
	switch {
	case err == nil:
		detail.Tiquetera = &tiquetera
	case !errors.Is(err, domain.ErrNotFound):
		return ClienteDetail{}, fmt.Errorf("getting active ticket book: %w", err)
	}
	return detail, nil
}

// List searches clients by name or phone. Page starts at 1.
func (s *ClienteService) List(ctx context.Context, q string, page, size int32) ([]domain.Cliente, int64, error) {
	items, total, err := s.repo.SearchClientes(ctx, q, size, (page-1)*size)
	if err != nil {
		return nil, 0, fmt.Errorf("searching clients: %w", err)
	}
	return items, total, nil
}

// ListConsumos returns a page of the client's history, newest first. Page starts at 1.
func (s *ClienteService) ListConsumos(ctx context.Context, id uuid.UUID, page, size int32) ([]domain.ConsumoEntry, int64, error) {
	if _, err := s.repo.GetCliente(ctx, id); err != nil {
		return nil, 0, fmt.Errorf("getting client: %w", err)
	}
	items, total, err := s.repo.ListConsumos(ctx, id, size, (page-1)*size)
	if err != nil {
		return nil, 0, fmt.Errorf("listing consumptions: %w", err)
	}
	return items, total, nil
}

// Update applies a partial update to a client.
func (s *ClienteService) Update(ctx context.Context, actor, id uuid.UUID, patch domain.ClientePatch) (domain.Cliente, error) {
	var updated domain.Cliente
	err := s.repo.InTx(ctx, func(tx ClienteRepository) error {
		current, txErr := tx.GetClienteForUpdate(ctx, id)
		if txErr != nil {
			return txErr
		}
		if current.ClienteData, txErr = patch.Apply(current.ClienteData).Normalize(); txErr != nil {
			return txErr
		}
		if patch.Activo != nil {
			current.Activo = *patch.Activo
		}
		if updated, txErr = tx.UpdateCliente(ctx, current); txErr != nil {
			return txErr
		}
		return auditCliente(ctx, tx, actor, domain.AuditClienteUpdated, domain.AuditEntityCliente, id)
	})
	if err != nil {
		return domain.Cliente{}, fmt.Errorf("updating client: %w", err)
	}
	return updated, nil
}

// Recharge replaces the client's active ticket book with a new one that also
// holds the previous balance, even if the previous book had expired.
func (s *ClienteService) Recharge(ctx context.Context, actor, id uuid.UUID, input domain.TiqueteraInput) (domain.Tiquetera, error) {
	fecha, err := input.Validate(s.clock.Now())
	if err != nil {
		return domain.Tiquetera{}, err
	}
	var created domain.Tiquetera
	err = s.repo.InTx(ctx, func(tx ClienteRepository) error {
		cliente, txErr := tx.GetClienteForUpdate(ctx, id)
		if txErr != nil {
			return txErr
		}
		if !cliente.Activo {
			return domain.ErrClienteInactive
		}
		previous, txErr := tx.DeactivateTiquetera(ctx, id)
		if txErr != nil {
			return txErr
		}
		total := input.Almuerzos + int(previous)
		if total > domain.MaxAlmuerzos {
			return domain.ValidationError{Message: "La recarga más el saldo actual no puede superar 500 almuerzos"}
		}
		if created, txErr = createTiquetera(ctx, tx, id, total, fecha, actor); txErr != nil {
			return txErr
		}
		return auditCliente(ctx, tx, actor, domain.AuditTiqueteraCreated, domain.AuditEntityTiquetera, created.ID)
	})
	if err != nil {
		return domain.Tiquetera{}, fmt.Errorf("recharging: %w", err)
	}
	return created, nil
}

// RegenerateQR invalidates the client's QR and returns the new payload.
func (s *ClienteService) RegenerateQR(ctx context.Context, actor, id uuid.UUID) (string, error) {
	payload, hash := domain.NewQRToken()
	err := s.repo.InTx(ctx, func(tx ClienteRepository) error {
		if txErr := tx.RotateQR(ctx, id, hash); txErr != nil {
			return txErr
		}
		return auditCliente(ctx, tx, actor, domain.AuditQRRegenerated, domain.AuditEntityCliente, id)
	})
	if err != nil {
		return "", fmt.Errorf("regenerating QR: %w", err)
	}
	return payload, nil
}

func createTiquetera(ctx context.Context, tx ClienteRepository, clienteID uuid.UUID, almuerzos int, fecha time.Time, actor uuid.UUID) (domain.Tiquetera, error) {
	count := int32(almuerzos) // #nosec G115 -- validated range 1..500
	return tx.CreateTiquetera(ctx, domain.Tiquetera{
		ClienteID:          clienteID,
		AlmuerzosComprados: count,
		Saldo:              count,
		FechaVencimiento:   fecha,
	}, actor)
}

func auditCliente(ctx context.Context, tx ClienteRepository, actor uuid.UUID, action, entity string, entityID uuid.UUID) error {
	return tx.AddAudit(ctx, domain.AuditEntry{UsuarioID: &actor, Accion: action, Entidad: entity, EntidadID: &entityID})
}
