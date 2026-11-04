package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"tiquetera/internal/domain"
	"tiquetera/internal/repository/postgres/sqlcgen"
	"tiquetera/internal/service"
)

const (
	pgUniqueViolation     = "23505"
	telefonoConstraint    = "clientes_telefono_key"
	likePatternEscapeRule = `\`
)

var likeEscaper = strings.NewReplacer(likePatternEscapeRule, `\\`, `%`, `\%`, `_`, `\_`)

// ClienteRepository implements service.ClienteRepository.
type ClienteRepository struct {
	pool *pgxpool.Pool
	q    *sqlcgen.Queries
}

// NewClienteRepository returns a pool-backed ClienteRepository.
func NewClienteRepository(pool *pgxpool.Pool) *ClienteRepository {
	return &ClienteRepository{pool: pool, q: sqlcgen.New(pool)}
}

// InTx runs fn with a repository bound to a new transaction.
func (r *ClienteRepository) InTx(ctx context.Context, fn func(service.ClienteRepository) error) error {
	return WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		return fn(&ClienteRepository{pool: r.pool, q: r.q.WithTx(tx)})
	})
}

// CreateCliente inserts a client or returns domain.ErrPhoneTaken.
func (r *ClienteRepository) CreateCliente(ctx context.Context, data domain.ClienteData, qrHash []byte, actor uuid.UUID) (domain.Cliente, error) {
	row, err := r.q.CreateCliente(ctx, sqlcgen.CreateClienteParams{
		Nombre:      data.Nombre,
		Telefono:    data.Telefono,
		Documento:   data.Documento,
		Email:       data.Email,
		QrTokenHash: qrHash,
		CreatedBy:   actor,
	})
	if err != nil {
		return domain.Cliente{}, mapWriteError(err)
	}
	return toCliente(row), nil
}

// GetCliente returns a client or domain.ErrNotFound.
func (r *ClienteRepository) GetCliente(ctx context.Context, id uuid.UUID) (domain.Cliente, error) {
	row, err := r.q.GetClienteByID(ctx, id)
	if err != nil {
		return domain.Cliente{}, mapNotFound(err)
	}
	return toCliente(row), nil
}

// GetClienteForUpdate locks and returns a client or domain.ErrNotFound.
func (r *ClienteRepository) GetClienteForUpdate(ctx context.Context, id uuid.UUID) (domain.Cliente, error) {
	row, err := r.q.GetClienteByIDForUpdate(ctx, id)
	if err != nil {
		return domain.Cliente{}, mapNotFound(err)
	}
	return toCliente(row), nil
}

// SearchClientes returns a page of clients matching q literally, and the total.
func (r *ClienteRepository) SearchClientes(ctx context.Context, q string, limit, offset int32) ([]domain.Cliente, int64, error) {
	pattern := likeEscaper.Replace(q)
	total, err := r.q.CountClientes(ctx, pattern)
	if err != nil {
		return nil, 0, fmt.Errorf("counting clients: %w", err)
	}
	rows, err := r.q.SearchClientes(ctx, sqlcgen.SearchClientesParams{Q: pattern, Size: limit, PageOffset: offset})
	if err != nil {
		return nil, 0, fmt.Errorf("searching clients: %w", err)
	}
	items := make([]domain.Cliente, 0, len(rows))
	for _, row := range rows {
		items = append(items, toCliente(row))
	}
	return items, total, nil
}

// UpdateCliente persists the client's editable fields or returns domain.ErrPhoneTaken.
func (r *ClienteRepository) UpdateCliente(ctx context.Context, c domain.Cliente) (domain.Cliente, error) {
	row, err := r.q.UpdateCliente(ctx, sqlcgen.UpdateClienteParams{
		ID:        c.ID,
		Nombre:    c.Nombre,
		Telefono:  c.Telefono,
		Documento: c.Documento,
		Email:     c.Email,
		Activo:    c.Activo,
	})
	if err != nil {
		return domain.Cliente{}, mapWriteError(err)
	}
	return toCliente(row), nil
}

// RotateQR replaces the QR hash, or returns domain.ErrNotFound.
func (r *ClienteRepository) RotateQR(ctx context.Context, id uuid.UUID, qrHash []byte) error {
	if _, err := r.q.RotateClienteQR(ctx, sqlcgen.RotateClienteQRParams{ID: id, QrTokenHash: qrHash}); err != nil {
		return mapNotFound(err)
	}
	return nil
}

// GetActiveTiquetera returns the active ticket book or domain.ErrNotFound.
func (r *ClienteRepository) GetActiveTiquetera(ctx context.Context, clienteID uuid.UUID) (domain.Tiquetera, error) {
	row, err := r.q.GetActiveTiqueteraByCliente(ctx, clienteID)
	if err != nil {
		return domain.Tiquetera{}, mapNotFound(err)
	}
	return toTiquetera(row), nil
}

// DeactivateTiquetera deactivates the client's active ticket book, if any, and
// returns its final balance. The row lock orders it against concurrent debits.
func (r *ClienteRepository) DeactivateTiquetera(ctx context.Context, clienteID uuid.UUID) (int32, error) {
	saldo, err := r.q.DeactivateActiveTiquetera(ctx, clienteID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("deactivating ticket book: %w", err)
	}
	return saldo, nil
}

// CreateTiquetera inserts an active ticket book.
func (r *ClienteRepository) CreateTiquetera(ctx context.Context, t domain.Tiquetera, actor uuid.UUID) (domain.Tiquetera, error) {
	row, err := r.q.CreateTiquetera(ctx, sqlcgen.CreateTiqueteraParams{
		ClienteID:          t.ClienteID,
		AlmuerzosComprados: t.AlmuerzosComprados,
		Saldo:              t.Saldo,
		FechaVencimiento:   t.FechaVencimiento,
		CreatedBy:          actor,
	})
	if err != nil {
		return domain.Tiquetera{}, fmt.Errorf("creating ticket book: %w", err)
	}
	return toTiquetera(row), nil
}

// ListConsumos returns a page of a client's consumptions, newest first, and the total.
func (r *ClienteRepository) ListConsumos(ctx context.Context, clienteID uuid.UUID, limit, offset int32) ([]domain.ConsumoEntry, int64, error) {
	total, err := r.q.CountConsumosByCliente(ctx, clienteID)
	if err != nil {
		return nil, 0, fmt.Errorf("counting consumptions: %w", err)
	}
	rows, err := r.q.ListConsumosByCliente(ctx, sqlcgen.ListConsumosByClienteParams{ClienteID: clienteID, Size: limit, PageOffset: offset})
	if err != nil {
		return nil, 0, fmt.Errorf("listing consumptions: %w", err)
	}
	items := make([]domain.ConsumoEntry, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.ConsumoEntry{
			ID:              row.ID,
			TiqueteraID:     row.TiqueteraID,
			ClienteID:       row.ClienteID,
			CajeroID:        row.CajeroID,
			FechaHora:       row.FechaHora.Time,
			SaldoResultante: row.SaldoResultante,
		})
	}
	return items, total, nil
}

// AddAudit inserts an audit_log row.
func (r *ClienteRepository) AddAudit(ctx context.Context, entry domain.AuditEntry) error {
	return addAudit(ctx, r.q, entry)
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation && pgErr.ConstraintName == telefonoConstraint {
		return domain.ErrPhoneTaken
	}
	return mapNotFound(err)
}

func toCliente(row sqlcgen.Cliente) domain.Cliente {
	return domain.Cliente{
		ClienteData: domain.ClienteData{
			Nombre:    row.Nombre,
			Telefono:  row.Telefono,
			Documento: row.Documento,
			Email:     row.Email,
		},
		ID:        row.ID,
		Activo:    row.Activo,
		QRVersion: row.QrVersion,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}

func toTiquetera(row sqlcgen.Tiquetera) domain.Tiquetera {
	return domain.Tiquetera{
		ID:                 row.ID,
		ClienteID:          row.ClienteID,
		AlmuerzosComprados: row.AlmuerzosComprados,
		Saldo:              row.Saldo,
		FechaVencimiento:   row.FechaVencimiento,
		Activa:             row.Activa,
		CreatedAt:          row.CreatedAt.Time,
	}
}
