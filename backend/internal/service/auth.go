// Package service implements the application use cases.
package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"tiquetera/internal/auth"
	"tiquetera/internal/domain"
	"tiquetera/internal/platform"
)

// maxConcurrentHashes bounds argon2id memory use (64 MB per verification).
const maxConcurrentHashes = 2

// AuthRepository is the persistence port for authentication.
type AuthRepository interface {
	InTx(ctx context.Context, fn func(AuthRepository) error) error
	GetUserByUsername(ctx context.Context, username string) (domain.User, error)
	GetUserByIDForUpdate(ctx context.Context, id uuid.UUID) (domain.User, error)
	UpdateLoginState(ctx context.Context, user domain.User) error
	CreateRefreshToken(ctx context.Context, token domain.RefreshToken) error
	GetRefreshTokenForUpdate(ctx context.Context, hash []byte) (domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id uuid.UUID, at time.Time) error
	RevokeRefreshFamily(ctx context.Context, familyID uuid.UUID, at time.Time) (int64, error)
	AddAudit(ctx context.Context, entry domain.AuditEntry) error
}

// AccessTokenIssuer signs access tokens.
type AccessTokenIssuer interface {
	IssueAccess(p domain.Principal) (string, error)
}

// AuthService implements login, refresh and logout.
type AuthService struct {
	repo      AuthRepository
	tokens    AccessTokenIssuer
	clock     platform.Clock
	dummyHash string
	hashSlots chan struct{}
}

// NewAuthService returns an AuthService.
func NewAuthService(repo AuthRepository, tokens AccessTokenIssuer, clock platform.Clock) (*AuthService, error) {
	dummy, err := platform.HashPassword(rand.Text())
	if err != nil {
		return nil, fmt.Errorf("creating dummy hash: %w", err)
	}
	return &AuthService{
		repo:      repo,
		tokens:    tokens,
		clock:     clock,
		dummyHash: dummy,
		hashSlots: make(chan struct{}, maxConcurrentHashes),
	}, nil
}

// txResult carries a business outcome out of a transaction that must commit,
// so that failed attempts and revocations persist even when the call fails.
type txResult struct {
	session domain.Session
	failure error
}

// Login authenticates a user. Every rejection returns ErrInvalidCredentials and
// performs a hash verification, so responses do not reveal whether the account
// exists, is inactive or is locked.
func (s *AuthService) Login(ctx context.Context, username, password string) (domain.Session, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.Session{}, s.rejectUnknownUser(ctx, password)
	}
	if err != nil {
		return domain.Session{}, fmt.Errorf("getting user: %w", err)
	}

	match, err := s.verifyPassword(ctx, password, user.PasswordHash)
	if err != nil {
		return domain.Session{}, err
	}

	var result txResult
	err = s.repo.InTx(ctx, func(tx AuthRepository) error {
		var txErr error
		result, txErr = s.finishLogin(ctx, tx, user.ID, match)
		return txErr
	})
	if err != nil {
		return domain.Session{}, fmt.Errorf("login transaction: %w", err)
	}
	return result.session, result.failure
}

func (s *AuthService) rejectUnknownUser(ctx context.Context, password string) error {
	if _, err := s.verifyPassword(ctx, password, s.dummyHash); err != nil {
		return err
	}
	if err := s.repo.AddAudit(ctx, domain.AuditEntry{Accion: domain.AuditLoginFailed, Entidad: domain.AuditEntityUsuario}); err != nil {
		return fmt.Errorf("auditing login: %w", err)
	}
	return domain.ErrInvalidCredentials
}

func (s *AuthService) verifyPassword(ctx context.Context, password, hash string) (bool, error) {
	select {
	case s.hashSlots <- struct{}{}:
		defer func() { <-s.hashSlots }()
	case <-ctx.Done():
		return false, fmt.Errorf("waiting for hash slot: %w", ctx.Err())
	}
	match, err := platform.VerifyPassword(password, hash)
	if err != nil {
		return false, fmt.Errorf("verifying password: %w", err)
	}
	return match, nil
}

func (s *AuthService) finishLogin(ctx context.Context, tx AuthRepository, id uuid.UUID, match bool) (txResult, error) {
	now := s.clock.Now()
	user, err := tx.GetUserByIDForUpdate(ctx, id)
	if err != nil {
		return txResult{}, fmt.Errorf("locking user: %w", err)
	}
	rejected := txResult{failure: domain.ErrInvalidCredentials}

	if !user.Activo || user.IsLocked(now) {
		return rejected, audit(ctx, tx, user.ID, domain.AuditLoginRejected)
	}
	if !match {
		return rejected, recordFailure(ctx, tx, user, now)
	}

	user.FailedLogins, user.LockedUntil = 0, nil
	if err := tx.UpdateLoginState(ctx, user); err != nil {
		return txResult{}, fmt.Errorf("resetting login state: %w", err)
	}
	session, err := s.newSession(ctx, tx, user, uuid.New(), now)
	if err != nil {
		return txResult{}, err
	}
	return txResult{session: session}, audit(ctx, tx, user.ID, domain.AuditLoginOK)
}

func recordFailure(ctx context.Context, tx AuthRepository, user domain.User, now time.Time) error {
	user = user.WithFailedLogin(now)
	if err := tx.UpdateLoginState(ctx, user); err != nil {
		return fmt.Errorf("recording failed login: %w", err)
	}
	if err := audit(ctx, tx, user.ID, domain.AuditLoginFailed); err != nil {
		return err
	}
	if user.IsLocked(now) {
		return audit(ctx, tx, user.ID, domain.AuditAccountLocked)
	}
	return nil
}

// Refresh rotates a refresh token. Presenting a revoked token is treated as
// theft: the whole family is revoked and the attempt is audited.
func (s *AuthService) Refresh(ctx context.Context, raw string) (domain.Session, error) {
	var result txResult
	err := s.repo.InTx(ctx, func(tx AuthRepository) error {
		var txErr error
		result, txErr = s.rotate(ctx, tx, raw)
		return txErr
	})
	if err != nil {
		return domain.Session{}, fmt.Errorf("refresh transaction: %w", err)
	}
	return result.session, result.failure
}

func (s *AuthService) rotate(ctx context.Context, tx AuthRepository, raw string) (txResult, error) {
	now := s.clock.Now()
	rejected := txResult{failure: domain.ErrInvalidRefreshToken}

	token, err := tx.GetRefreshTokenForUpdate(ctx, auth.HashRefreshToken(raw))
	if errors.Is(err, domain.ErrNotFound) {
		return rejected, nil
	}
	if err != nil {
		return txResult{}, fmt.Errorf("getting refresh token: %w", err)
	}
	if token.RevokedAt != nil {
		return rejected, revokeReusedFamily(ctx, tx, token, now)
	}
	if !token.ExpiresAt.After(now) {
		return rejected, nil
	}

	user, err := tx.GetUserByIDForUpdate(ctx, token.UsuarioID)
	if err != nil {
		return txResult{}, fmt.Errorf("locking user: %w", err)
	}
	if !user.Activo {
		if _, err := tx.RevokeRefreshFamily(ctx, token.FamilyID, now); err != nil {
			return txResult{}, fmt.Errorf("revoking family: %w", err)
		}
		return rejected, nil
	}

	if err := tx.RevokeRefreshToken(ctx, token.ID, now); err != nil {
		return txResult{}, fmt.Errorf("revoking refresh token: %w", err)
	}
	session, err := s.newSession(ctx, tx, user, token.FamilyID, now)
	if err != nil {
		return txResult{}, err
	}
	return txResult{session: session}, nil
}

// revokeReusedFamily audits only when tokens were still active, so replaying a
// token from an already closed session does not raise a false theft alert.
func revokeReusedFamily(ctx context.Context, tx AuthRepository, token domain.RefreshToken, now time.Time) error {
	revoked, err := tx.RevokeRefreshFamily(ctx, token.FamilyID, now)
	if err != nil {
		return fmt.Errorf("revoking family: %w", err)
	}
	if revoked == 0 {
		return nil
	}
	return audit(ctx, tx, token.UsuarioID, domain.AuditRefreshReused)
}

// Logout revokes the family of a refresh token owned by the principal. Unknown
// tokens and tokens of other users are ignored, so it reveals nothing about them.
func (s *AuthService) Logout(ctx context.Context, p domain.Principal, raw string) error {
	err := s.repo.InTx(ctx, func(tx AuthRepository) error {
		token, err := tx.GetRefreshTokenForUpdate(ctx, auth.HashRefreshToken(raw))
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("getting refresh token: %w", err)
		}
		if token.UsuarioID != p.UserID {
			return nil
		}
		if _, err := tx.RevokeRefreshFamily(ctx, token.FamilyID, s.clock.Now()); err != nil {
			return fmt.Errorf("revoking family: %w", err)
		}
		return audit(ctx, tx, p.UserID, domain.AuditLogout)
	})
	if err != nil {
		return fmt.Errorf("logout transaction: %w", err)
	}
	return nil
}

func (s *AuthService) newSession(ctx context.Context, tx AuthRepository, user domain.User, family uuid.UUID, now time.Time) (domain.Session, error) {
	raw, hash, err := auth.NewRefreshToken()
	if err != nil {
		return domain.Session{}, fmt.Errorf("creating refresh token: %w", err)
	}
	err = tx.CreateRefreshToken(ctx, domain.RefreshToken{
		UsuarioID: user.ID,
		TokenHash: hash,
		FamilyID:  family,
		ExpiresAt: now.Add(auth.RefreshTokenTTL),
	})
	if err != nil {
		return domain.Session{}, fmt.Errorf("storing refresh token: %w", err)
	}
	access, err := s.tokens.IssueAccess(domain.Principal{UserID: user.ID, Rol: user.Rol})
	if err != nil {
		return domain.Session{}, fmt.Errorf("issuing access token: %w", err)
	}
	return domain.Session{
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresIn:    int(auth.AccessTokenTTL.Seconds()),
		User:         user,
	}, nil
}

func audit(ctx context.Context, tx AuthRepository, userID uuid.UUID, action string) error {
	err := tx.AddAudit(ctx, domain.AuditEntry{
		UsuarioID: &userID,
		Accion:    action,
		Entidad:   domain.AuditEntityUsuario,
		EntidadID: &userID,
	})
	if err != nil {
		return fmt.Errorf("auditing %s: %w", action, err)
	}
	return nil
}
