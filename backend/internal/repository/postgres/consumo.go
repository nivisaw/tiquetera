package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"tiquetera/internal/domain"
	"tiquetera/internal/repository/postgres/sqlcgen"
	"tiquetera/internal/service"
)

const idempotencyConstraint = "consumos_idempotency_key_key"

// ConsumoRepository implements service.ConsumoRepository.
type ConsumoRepository struct {
	pool *pgxpool.Pool
	q    *sqlcgen.Queries
}

// NewConsumoRepository returns a pool-backed ConsumoRepository.
func NewConsumoRepository(pool *pgxpool.Pool) *ConsumoRepository {
	return &ConsumoRepository{pool: pool, q: sqlcgen.New(pool)}
}

// InTx runs fn with a repository bound to a new transaction.
func (r *ConsumoRepository) InTx(ctx context.Context, fn func(service.ConsumoRepository) error) error {
	return WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		return fn(&ConsumoRepository{pool: r.pool, q: r.q.WithTx(tx)})
	})
}

// GetConsumoResultado returns the result stored for a key or domain.ErrNotFound.
func (r *ConsumoRepository) GetConsumoResultado(ctx context.Context, key uuid.UUID) (domain.ConsumoResult, error) {
	row, err := r.q.GetConsumoResultadoByKey(ctx, key)
	if err != nil {
		return domain.ConsumoResult{}, mapNotFound(err)
	}
	return domain.ConsumoResult{
		CajeroID:         row.CajeroID,
		ClienteNombre:    row.Nombre,
		SaldoRestante:    row.SaldoResultante,
		FechaVencimiento: row.FechaVencimiento,
		FechaHora:        row.FechaHora.Time,
	}, nil
}

// GetClienteByQRHash returns the QR's client or domain.ErrNotFound.
func (r *ConsumoRepository) GetClienteByQRHash(ctx context.Context, hash []byte) (domain.Cliente, error) {
	row, err := r.q.GetClienteByQRHash(ctx, hash)
	if err != nil {
		return domain.Cliente{}, mapNotFound(err)
	}
	return toCliente(row), nil
}

// GetCliente returns a client or domain.ErrNotFound.
func (r *ConsumoRepository) GetCliente(ctx context.Context, id uuid.UUID) (domain.Cliente, error) {
	row, err := r.q.GetClienteByID(ctx, id)
	if err != nil {
		return domain.Cliente{}, mapNotFound(err)
	}
	return toCliente(row), nil
}

// GetActiveTiquetera returns the active ticket book or domain.ErrNotFound.
func (r *ConsumoRepository) GetActiveTiquetera(ctx context.Context, clienteID uuid.UUID) (domain.Tiquetera, error) {
	row, err := r.q.GetActiveTiqueteraByCliente(ctx, clienteID)
	if err != nil {
		return domain.Tiquetera{}, mapNotFound(err)
	}
	return toTiquetera(row), nil
}

// DebitTiquetera atomically takes one lunch from an eligible ticket book, or
// returns domain.ErrNotFound if there is none.
func (r *ConsumoRepository) DebitTiquetera(ctx context.Context, clienteID uuid.UUID, today time.Time) (domain.Tiquetera, error) {
	row, err := r.q.DebitTiquetera(ctx, sqlcgen.DebitTiqueteraParams{ClienteID: clienteID, Hoy: today})
	if err != nil {
		return domain.Tiquetera{}, mapNotFound(err)
	}
	return toTiquetera(row), nil
}

// CreateConsumo stores a debit and returns its time, or
// domain.ErrDuplicateIdempotencyKey if the key was already used.
func (r *ConsumoRepository) CreateConsumo(ctx context.Context, c domain.Consumo) (time.Time, error) {
	fechaHora, err := r.q.CreateConsumo(ctx, sqlcgen.CreateConsumoParams{
		TiqueteraID:     c.TiqueteraID,
		ClienteID:       c.ClienteID,
		CajeroID:        c.CajeroID,
		SaldoResultante: c.SaldoResultante,
		IdempotencyKey:  c.IdempotencyKey,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation && pgErr.ConstraintName == idempotencyConstraint {
			return time.Time{}, domain.ErrDuplicateIdempotencyKey
		}
		return time.Time{}, fmt.Errorf("creating consumption: %w", err)
	}
	return fechaHora.Time, nil
}

// AddRejection stores a rejected attempt.
func (r *ConsumoRepository) AddRejection(ctx context.Context, cajero uuid.UUID, cliente *uuid.UUID, code domain.RejectionCode) error {
	err := r.q.CreateConsumoRechazado(ctx, sqlcgen.CreateConsumoRechazadoParams{
		CajeroID:  cajero,
		ClienteID: cliente,
		Motivo:    string(code),
	})
	if err != nil {
		return fmt.Errorf("inserting rejection: %w", err)
	}
	return nil
}
