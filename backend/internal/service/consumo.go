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

// maxDebitAttempts covers a recharge replacing the ticket book during a debit:
// the debit finds no eligible book, yet the diagnosis sees the new one.
const maxDebitAttempts = 3

var errDebitRaced = errors.New("ticket book changed during debit")

// ConsumoRepository is the persistence port for consumption.
type ConsumoRepository interface {
	InTx(ctx context.Context, fn func(ConsumoRepository) error) error
	GetConsumoResultado(ctx context.Context, key uuid.UUID) (domain.ConsumoResult, error)
	GetClienteByQRHash(ctx context.Context, hash []byte) (domain.Cliente, error)
	GetCliente(ctx context.Context, id uuid.UUID) (domain.Cliente, error)
	GetActiveTiquetera(ctx context.Context, clienteID uuid.UUID) (domain.Tiquetera, error)
	DebitTiquetera(ctx context.Context, clienteID uuid.UUID, today time.Time) (domain.Tiquetera, error)
	CreateConsumo(ctx context.Context, c domain.Consumo) (time.Time, error)
	AddRejection(ctx context.Context, cajero uuid.UUID, cliente *uuid.UUID, code domain.RejectionCode) error
}

// ConsumoService debits lunches.
type ConsumoService struct {
	repo  ConsumoRepository
	clock platform.Clock
}

// NewConsumoService returns a ConsumoService.
func NewConsumoService(repo ConsumoRepository, clock platform.Clock) *ConsumoService {
	return &ConsumoService{repo: repo, clock: clock}
}

// Consume debits one lunch for the QR's client. A repeated key returns the
// original result without debiting again. Rejections are recorded and returned
// as domain.ConsumoRejection; they debit nothing, so a retry is evaluated anew.
func (s *ConsumoService) Consume(ctx context.Context, cajero, key uuid.UUID, qr string) (domain.ConsumoResult, error) {
	result, err := s.replay(ctx, cajero, key)
	if !errors.Is(err, domain.ErrNotFound) {
		return result, err
	}

	token, ok := domain.ParseQRPayload(qr)
	if !ok {
		return domain.ConsumoResult{}, s.reject(ctx, s.repo, cajero, nil, domain.NewConsumoRejection(domain.RejectQRInvalid))
	}
	cliente, err := s.repo.GetClienteByQRHash(ctx, domain.HashQRToken(token))
	if errors.Is(err, domain.ErrNotFound) {
		return domain.ConsumoResult{}, s.reject(ctx, s.repo, cajero, nil, domain.NewConsumoRejection(domain.RejectClienteNotFound))
	}
	if err != nil {
		return domain.ConsumoResult{}, fmt.Errorf("finding client by QR: %w", err)
	}

	for range maxDebitAttempts {
		result, err = s.attemptDebit(ctx, cajero, key, cliente)
		switch {
		case errors.Is(err, errDebitRaced):
			continue
		case errors.Is(err, domain.ErrDuplicateIdempotencyKey):
			return s.replay(ctx, cajero, key)
		default:
			return result, err
		}
	}
	return domain.ConsumoResult{}, fmt.Errorf("debiting: %w", err)
}

// replay returns the stored result for a key, or domain.ErrNotFound if unused.
func (s *ConsumoService) replay(ctx context.Context, cajero, key uuid.UUID) (domain.ConsumoResult, error) {
	result, err := s.repo.GetConsumoResultado(ctx, key)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.ConsumoResult{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.ConsumoResult{}, fmt.Errorf("finding previous consumption: %w", err)
	}
	if result.CajeroID != cajero {
		return domain.ConsumoResult{}, domain.ErrIdempotencyConflict
	}
	return result, nil
}

// attemptDebit runs one transaction. A rejection commits, so its record persists.
func (s *ConsumoService) attemptDebit(ctx context.Context, cajero, key uuid.UUID, cliente domain.Cliente) (domain.ConsumoResult, error) {
	today := domain.TodayInBogota(s.clock.Now())
	var result domain.ConsumoResult
	var rejection error
	err := s.repo.InTx(ctx, func(tx ConsumoRepository) error {
		tiquetera, txErr := tx.DebitTiquetera(ctx, cliente.ID, today)
		if errors.Is(txErr, domain.ErrNotFound) {
			rejection, txErr = s.diagnose(ctx, tx, cajero, cliente.ID, today)
			return txErr
		}
		if txErr != nil {
			return fmt.Errorf("debiting ticket book: %w", txErr)
		}
		fechaHora, txErr := tx.CreateConsumo(ctx, domain.Consumo{
			TiqueteraID:     tiquetera.ID,
			ClienteID:       cliente.ID,
			CajeroID:        cajero,
			SaldoResultante: tiquetera.Saldo,
			IdempotencyKey:  key,
		})
		if txErr != nil {
			return txErr
		}
		result = domain.ConsumoResult{
			CajeroID:         cajero,
			ClienteNombre:    cliente.Nombre,
			SaldoRestante:    tiquetera.Saldo,
			FechaVencimiento: tiquetera.FechaVencimiento,
			FechaHora:        fechaHora,
		}
		return nil
	})
	if err != nil {
		return domain.ConsumoResult{}, fmt.Errorf("consumption transaction: %w", err)
	}
	return result, rejection
}

// diagnose records why the debit did not apply and returns it as the rejection,
// or errDebitRaced when nothing prevents the debit any more.
func (s *ConsumoService) diagnose(ctx context.Context, tx ConsumoRepository, cajero, clienteID uuid.UUID, today time.Time) (rejection, err error) {
	cliente, err := tx.GetCliente(ctx, clienteID)
	if err != nil {
		return nil, fmt.Errorf("reading client: %w", err)
	}
	var active *domain.Tiquetera
	tiquetera, err := tx.GetActiveTiquetera(ctx, clienteID)
	switch {
	case err == nil:
		active = &tiquetera
	case !errors.Is(err, domain.ErrNotFound):
		return nil, fmt.Errorf("reading ticket book: %w", err)
	}
	found, rejected := domain.DiagnoseRejection(cliente, active, today)
	if !rejected {
		return errDebitRaced, nil
	}
	return s.reject(ctx, tx, cajero, &clienteID, found), nil
}

// reject stores the rejection and returns it, or the storage failure.
func (s *ConsumoService) reject(ctx context.Context, repo ConsumoRepository, cajero uuid.UUID, cliente *uuid.UUID, rejection domain.ConsumoRejection) error {
	if err := repo.AddRejection(ctx, cajero, cliente, rejection.Code); err != nil {
		return fmt.Errorf("recording rejection: %w", err)
	}
	return rejection
}
