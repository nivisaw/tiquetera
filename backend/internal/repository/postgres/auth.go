// Package postgres implements the persistence ports on PostgreSQL.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"tiquetera/internal/domain"
	"tiquetera/internal/repository/postgres/sqlcgen"
	"tiquetera/internal/service"
)

// WithTx runs fn in a transaction: commits on nil, rolls back on error.
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	if err := pgx.BeginFunc(ctx, pool, fn); err != nil {
		return fmt.Errorf("transaction: %w", err)
	}
	return nil
}

// AuthRepository implements service.AuthRepository.
type AuthRepository struct {
	pool *pgxpool.Pool
	q    *sqlcgen.Queries
}

// NewAuthRepository returns a pool-backed AuthRepository.
func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool, q: sqlcgen.New(pool)}
}

// InTx runs fn with a repository bound to a new transaction.
func (r *AuthRepository) InTx(ctx context.Context, fn func(service.AuthRepository) error) error {
	return WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		return fn(&AuthRepository{pool: r.pool, q: r.q.WithTx(tx)})
	})
}

// GetUserByUsername returns a user or domain.ErrNotFound.
func (r *AuthRepository) GetUserByUsername(ctx context.Context, username string) (domain.User, error) {
	row, err := r.q.GetUsuarioByUsername(ctx, username)
	if err != nil {
		return domain.User{}, mapNotFound(err)
	}
	return toUser(row), nil
}

// GetUserByIDForUpdate locks and returns a user or domain.ErrNotFound.
func (r *AuthRepository) GetUserByIDForUpdate(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row, err := r.q.GetUsuarioByIDForUpdate(ctx, id)
	if err != nil {
		return domain.User{}, mapNotFound(err)
	}
	return toUser(row), nil
}

// UpdateLoginState persists the failed-login counter and lock.
func (r *AuthRepository) UpdateLoginState(ctx context.Context, user domain.User) error {
	err := r.q.UpdateUsuarioLoginState(ctx, sqlcgen.UpdateUsuarioLoginStateParams{
		ID:           user.ID,
		FailedLogins: user.FailedLogins,
		LockedUntil:  fromTimePtr(user.LockedUntil),
	})
	if err != nil {
		return fmt.Errorf("updating login state: %w", err)
	}
	return nil
}

// CreateRefreshToken stores a refresh token hash.
func (r *AuthRepository) CreateRefreshToken(ctx context.Context, token domain.RefreshToken) error {
	err := r.q.CreateRefreshToken(ctx, sqlcgen.CreateRefreshTokenParams{
		UsuarioID: token.UsuarioID,
		TokenHash: token.TokenHash,
		FamilyID:  token.FamilyID,
		ExpiresAt: fromTime(token.ExpiresAt),
	})
	if err != nil {
		return fmt.Errorf("creating refresh token: %w", err)
	}
	return nil
}

// GetRefreshTokenForUpdate locks and returns a token or domain.ErrNotFound.
func (r *AuthRepository) GetRefreshTokenForUpdate(ctx context.Context, hash []byte) (domain.RefreshToken, error) {
	row, err := r.q.GetRefreshTokenByHashForUpdate(ctx, hash)
	if err != nil {
		return domain.RefreshToken{}, mapNotFound(err)
	}
	return domain.RefreshToken{
		ID:        row.ID,
		UsuarioID: row.UsuarioID,
		TokenHash: row.TokenHash,
		FamilyID:  row.FamilyID,
		ExpiresAt: row.ExpiresAt.Time,
		RevokedAt: toTimePtr(row.RevokedAt),
	}, nil
}

// RevokeRefreshToken revokes one token.
func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID, at time.Time) error {
	err := r.q.RevokeRefreshToken(ctx, sqlcgen.RevokeRefreshTokenParams{ID: id, RevokedAt: fromTime(at)})
	if err != nil {
		return fmt.Errorf("revoking refresh token: %w", err)
	}
	return nil
}

// RevokeRefreshFamily revokes the active tokens of a family and returns how many.
func (r *AuthRepository) RevokeRefreshFamily(ctx context.Context, familyID uuid.UUID, at time.Time) (int64, error) {
	n, err := r.q.RevokeRefreshFamily(ctx, sqlcgen.RevokeRefreshFamilyParams{FamilyID: familyID, RevokedAt: fromTime(at)})
	if err != nil {
		return 0, fmt.Errorf("revoking refresh family: %w", err)
	}
	return n, nil
}

// AddAudit inserts an audit_log row.
func (r *AuthRepository) AddAudit(ctx context.Context, entry domain.AuditEntry) error {
	return addAudit(ctx, r.q, entry)
}

func addAudit(ctx context.Context, q *sqlcgen.Queries, entry domain.AuditEntry) error {
	err := q.CreateAuditLog(ctx, sqlcgen.CreateAuditLogParams{
		UsuarioID: entry.UsuarioID,
		Accion:    entry.Accion,
		Entidad:   entry.Entidad,
		EntidadID: entry.EntidadID,
		Datos:     []byte("{}"),
	})
	if err != nil {
		return fmt.Errorf("inserting audit log: %w", err)
	}
	return nil
}

func toUser(row sqlcgen.Usuario) domain.User {
	return domain.User{
		ID:           row.ID,
		Nombre:       row.Nombre,
		PasswordHash: row.PasswordHash,
		Rol:          domain.Role(row.Rol),
		Activo:       row.Activo,
		FailedLogins: row.FailedLogins,
		LockedUntil:  toTimePtr(row.LockedUntil),
	}
}

func mapNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return fmt.Errorf("query: %w", err)
}

func fromTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func fromTimePtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return fromTime(*t)
}

func toTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
