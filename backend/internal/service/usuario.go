package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"tiquetera/internal/domain"
	"tiquetera/internal/platform"
)

// UsuarioRepository is the persistence port for account management.
type UsuarioRepository interface {
	InTx(ctx context.Context, fn func(UsuarioRepository) error) error
	CreateUser(ctx context.Context, user domain.NewUser, passwordHash string) (domain.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (domain.User, error)
	GetUserForUpdate(ctx context.Context, id uuid.UUID) (domain.User, error)
	ListUsers(ctx context.Context, limit, offset int32) ([]domain.User, int64, error)
	LockActiveAdmins(ctx context.Context) (int, error)
	UpdateUser(ctx context.Context, user domain.User) (domain.User, error)
	SetPassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	RevokeUserSessions(ctx context.Context, id uuid.UUID, at time.Time) error
	AddAudit(ctx context.Context, entry domain.AuditEntry) error
}

// UsuarioService implements account management for admins.
type UsuarioService struct {
	repo      UsuarioRepository
	clock     platform.Clock
	hashSlots chan struct{}
}

// NewUsuarioService returns a UsuarioService.
func NewUsuarioService(repo UsuarioRepository, clock platform.Clock) *UsuarioService {
	return &UsuarioService{repo: repo, clock: clock, hashSlots: make(chan struct{}, maxConcurrentHashes)}
}

// Create registers an account.
func (s *UsuarioService) Create(ctx context.Context, actor uuid.UUID, input domain.NewUser) (domain.User, error) {
	input, err := input.Normalize()
	if err != nil {
		return domain.User{}, err
	}
	hash, err := s.hashPassword(ctx, input.Password)
	if err != nil {
		return domain.User{}, err
	}
	var created domain.User
	err = s.repo.InTx(ctx, func(tx UsuarioRepository) error {
		var txErr error
		if created, txErr = tx.CreateUser(ctx, input, hash); txErr != nil {
			return txErr
		}
		return auditUsuario(ctx, tx, actor, domain.AuditUsuarioCreated, created.ID)
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("creating user: %w", err)
	}
	return created, nil
}

// Get returns an account.
func (s *UsuarioService) Get(ctx context.Context, id uuid.UUID) (domain.User, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("getting user: %w", err)
	}
	return user, nil
}

// List returns a page of accounts. Page starts at 1.
func (s *UsuarioService) List(ctx context.Context, page, size int32) ([]domain.User, int64, error) {
	items, total, err := s.repo.ListUsers(ctx, size, (page-1)*size)
	if err != nil {
		return nil, 0, fmt.Errorf("listing users: %w", err)
	}
	return items, total, nil
}

// Update applies a partial update. Active admins are locked first, so concurrent
// changes cannot leave the system without one. Losing access or changing role
// closes the account's sessions.
func (s *UsuarioService) Update(ctx context.Context, actor, id uuid.UUID, patch domain.UserPatch) (domain.User, error) {
	patch, err := patch.Normalize()
	if err != nil {
		return domain.User{}, err
	}
	var updated domain.User
	err = s.repo.InTx(ctx, func(tx UsuarioRepository) error {
		var txErr error
		updated, txErr = s.applyPatch(ctx, tx, id, patch)
		if txErr != nil {
			return txErr
		}
		return auditUsuario(ctx, tx, actor, domain.AuditUsuarioUpdated, id)
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("updating user: %w", err)
	}
	return updated, nil
}

func (s *UsuarioService) applyPatch(ctx context.Context, tx UsuarioRepository, id uuid.UUID, patch domain.UserPatch) (domain.User, error) {
	activeAdmins := 0
	if patch.ChangesAccess() {
		var err error
		if activeAdmins, err = tx.LockActiveAdmins(ctx); err != nil {
			return domain.User{}, err
		}
	}
	current, err := tx.GetUserForUpdate(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	next := patch.Apply(current)
	if current.IsActiveAdmin() && !next.IsActiveAdmin() && activeAdmins <= 1 {
		return domain.User{}, domain.ErrLastAdmin
	}
	updated, err := tx.UpdateUser(ctx, next)
	if err != nil {
		return domain.User{}, err
	}
	if current.Rol != next.Rol || (current.Activo && !next.Activo) {
		if err := tx.RevokeUserSessions(ctx, id, s.clock.Now()); err != nil {
			return domain.User{}, err
		}
	}
	return updated, nil
}

// ResetPassword replaces the password, clears any lock and closes all sessions.
func (s *UsuarioService) ResetPassword(ctx context.Context, actor, id uuid.UUID, password string) error {
	if err := domain.ValidatePassword(password); err != nil {
		return err
	}
	hash, err := s.hashPassword(ctx, password)
	if err != nil {
		return err
	}
	err = s.repo.InTx(ctx, func(tx UsuarioRepository) error {
		if _, txErr := tx.GetUserForUpdate(ctx, id); txErr != nil {
			return txErr
		}
		if txErr := tx.SetPassword(ctx, id, hash); txErr != nil {
			return txErr
		}
		if txErr := tx.RevokeUserSessions(ctx, id, s.clock.Now()); txErr != nil {
			return txErr
		}
		return auditUsuario(ctx, tx, actor, domain.AuditPasswordChanged, id)
	})
	if err != nil {
		return fmt.Errorf("resetting password: %w", err)
	}
	return nil
}

func (s *UsuarioService) hashPassword(ctx context.Context, password string) (string, error) {
	select {
	case s.hashSlots <- struct{}{}:
		defer func() { <-s.hashSlots }()
	case <-ctx.Done():
		return "", fmt.Errorf("waiting for hash slot: %w", ctx.Err())
	}
	hash, err := platform.HashPassword(password)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return hash, nil
}

func auditUsuario(ctx context.Context, tx UsuarioRepository, actor uuid.UUID, action string, id uuid.UUID) error {
	return tx.AddAudit(ctx, domain.AuditEntry{UsuarioID: &actor, Accion: action, Entidad: domain.AuditEntityUsuario, EntidadID: &id})
}
