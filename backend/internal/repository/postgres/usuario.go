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

const usernameConstraint = "usuarios_username_key"

// UsuarioRepository implements service.UsuarioRepository.
type UsuarioRepository struct {
	pool *pgxpool.Pool
	q    *sqlcgen.Queries
}

// NewUsuarioRepository returns a pool-backed UsuarioRepository.
func NewUsuarioRepository(pool *pgxpool.Pool) *UsuarioRepository {
	return &UsuarioRepository{pool: pool, q: sqlcgen.New(pool)}
}

// InTx runs fn with a repository bound to a new transaction.
func (r *UsuarioRepository) InTx(ctx context.Context, fn func(service.UsuarioRepository) error) error {
	return WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		return fn(&UsuarioRepository{pool: r.pool, q: r.q.WithTx(tx)})
	})
}

// CreateUser inserts an account or returns domain.ErrUsernameTaken.
func (r *UsuarioRepository) CreateUser(ctx context.Context, user domain.NewUser, passwordHash string) (domain.User, error) {
	row, err := r.q.CreateUsuario(ctx, sqlcgen.CreateUsuarioParams{
		Username:     user.Username,
		PasswordHash: passwordHash,
		Nombre:       user.Nombre,
		Rol:          string(user.Rol),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation && pgErr.ConstraintName == usernameConstraint {
			return domain.User{}, domain.ErrUsernameTaken
		}
		return domain.User{}, fmt.Errorf("creating user: %w", err)
	}
	return toUser(row), nil
}

// GetUser returns an account or domain.ErrNotFound.
func (r *UsuarioRepository) GetUser(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row, err := r.q.GetUsuarioByID(ctx, id)
	if err != nil {
		return domain.User{}, mapNotFound(err)
	}
	return toUser(row), nil
}

// GetUserForUpdate locks and returns an account or domain.ErrNotFound.
func (r *UsuarioRepository) GetUserForUpdate(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row, err := r.q.GetUsuarioByIDForUpdate(ctx, id)
	if err != nil {
		return domain.User{}, mapNotFound(err)
	}
	return toUser(row), nil
}

// ListUsers returns a page of accounts and the total.
func (r *UsuarioRepository) ListUsers(ctx context.Context, limit, offset int32) ([]domain.User, int64, error) {
	total, err := r.q.CountUsuarios(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("counting users: %w", err)
	}
	rows, err := r.q.ListUsuarios(ctx, sqlcgen.ListUsuariosParams{Size: limit, PageOffset: offset})
	if err != nil {
		return nil, 0, fmt.Errorf("listing users: %w", err)
	}
	items := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		items = append(items, toUser(row))
	}
	return items, total, nil
}

// LockActiveAdmins locks the active admins until the transaction ends and
// returns how many there are.
func (r *UsuarioRepository) LockActiveAdmins(ctx context.Context) (int, error) {
	ids, err := r.q.LockActiveAdmins(ctx)
	if err != nil {
		return 0, fmt.Errorf("locking active admins: %w", err)
	}
	return len(ids), nil
}

// UpdateUser persists an account's editable fields.
func (r *UsuarioRepository) UpdateUser(ctx context.Context, user domain.User) (domain.User, error) {
	row, err := r.q.UpdateUsuario(ctx, sqlcgen.UpdateUsuarioParams{
		ID:     user.ID,
		Nombre: user.Nombre,
		Rol:    string(user.Rol),
		Activo: user.Activo,
	})
	if err != nil {
		return domain.User{}, mapNotFound(err)
	}
	return toUser(row), nil
}

// SetPassword replaces the hash and clears the lock, or returns domain.ErrNotFound.
func (r *UsuarioRepository) SetPassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	n, err := r.q.UpdateUsuarioPassword(ctx, sqlcgen.UpdateUsuarioPasswordParams{ID: id, PasswordHash: passwordHash})
	if err != nil {
		return fmt.Errorf("updating password: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// RevokeUserSessions revokes every active refresh token of an account.
func (r *UsuarioRepository) RevokeUserSessions(ctx context.Context, id uuid.UUID, at time.Time) error {
	err := r.q.RevokeRefreshTokensByUsuario(ctx, sqlcgen.RevokeRefreshTokensByUsuarioParams{UsuarioID: id, RevokedAt: fromTime(at)})
	if err != nil {
		return fmt.Errorf("revoking user sessions: %w", err)
	}
	return nil
}

// AddAudit inserts an audit_log row.
func (r *UsuarioRepository) AddAudit(ctx context.Context, entry domain.AuditEntry) error {
	return addAudit(ctx, r.q, entry)
}
