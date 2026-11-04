package service

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tiquetera/internal/auth"
	"tiquetera/internal/domain"
	"tiquetera/internal/platform"
)

const testPassword = "clave-correcta-123"

type fakeClock struct{ t time.Time }

func (c *fakeClock) Now() time.Time { return c.t }

type fakeIssuer struct{ err error }

func (f fakeIssuer) IssueAccess(p domain.Principal) (string, error) {
	return "access-" + p.UserID.String(), f.err
}

type fakeRepo struct {
	user    domain.User
	tokens  []*domain.RefreshToken
	audits  []domain.AuditEntry
	failGet error
}

func (r *fakeRepo) InTx(_ context.Context, fn func(AuthRepository) error) error { return fn(r) }

func (r *fakeRepo) GetUserByUsername(_ context.Context, username string) (domain.User, error) {
	if r.failGet != nil {
		return domain.User{}, r.failGet
	}
	if username != "ana" {
		return domain.User{}, domain.ErrNotFound
	}
	return r.user, nil
}

func (r *fakeRepo) GetUserByIDForUpdate(context.Context, uuid.UUID) (domain.User, error) {
	return r.user, nil
}

func (r *fakeRepo) UpdateLoginState(_ context.Context, u domain.User) error {
	r.user.FailedLogins, r.user.LockedUntil = u.FailedLogins, u.LockedUntil
	return nil
}

func (r *fakeRepo) CreateRefreshToken(_ context.Context, t domain.RefreshToken) error {
	t.ID = uuid.New()
	r.tokens = append(r.tokens, &t)
	return nil
}

func (r *fakeRepo) GetRefreshTokenForUpdate(_ context.Context, hash []byte) (domain.RefreshToken, error) {
	for _, t := range r.tokens {
		if bytes.Equal(t.TokenHash, hash) {
			return *t, nil
		}
	}
	return domain.RefreshToken{}, domain.ErrNotFound
}

func (r *fakeRepo) RevokeRefreshToken(_ context.Context, id uuid.UUID, at time.Time) error {
	for _, t := range r.tokens {
		if t.ID == id {
			t.RevokedAt = &at
		}
	}
	return nil
}

func (r *fakeRepo) RevokeRefreshFamily(_ context.Context, family uuid.UUID, at time.Time) (int64, error) {
	var n int64
	for _, t := range r.tokens {
		if t.FamilyID == family && t.RevokedAt == nil {
			t.RevokedAt = &at
			n++
		}
	}
	return n, nil
}

func (r *fakeRepo) AddAudit(_ context.Context, e domain.AuditEntry) error {
	r.audits = append(r.audits, e)
	return nil
}

func (r *fakeRepo) actions() []string {
	out := make([]string, 0, len(r.audits))
	for _, a := range r.audits {
		out = append(out, a.Accion)
	}
	return out
}

func newService(t *testing.T) (*AuthService, *fakeRepo, *fakeClock) {
	t.Helper()
	hash, err := platform.HashPassword(testPassword)
	require.NoError(t, err)
	repo := &fakeRepo{user: domain.User{ID: uuid.New(), Nombre: "Ana", PasswordHash: hash, Rol: domain.RoleCajero, Activo: true}}
	clock := &fakeClock{t: time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)}
	svc, err := NewAuthService(repo, fakeIssuer{}, clock)
	require.NoError(t, err)
	return svc, repo, clock
}

func TestLogin_ValidCredentials_ReturnsSessionAndStoresOnlyHash(t *testing.T) {
	svc, repo, clock := newService(t)
	repo.user.FailedLogins = 3

	s, err := svc.Login(t.Context(), "ana", testPassword)
	require.NoError(t, err)
	assert.Equal(t, 900, s.ExpiresIn)
	assert.Equal(t, "Ana", s.User.Nombre)
	assert.Equal(t, int32(0), repo.user.FailedLogins)
	require.Len(t, repo.tokens, 1)
	assert.Equal(t, auth.HashRefreshToken(s.RefreshToken), repo.tokens[0].TokenHash)
	assert.Equal(t, clock.t.Add(30*24*time.Hour), repo.tokens[0].ExpiresAt)
	assert.Equal(t, []string{domain.AuditLoginOK}, repo.actions())
}

func TestLogin_WrongPassword_CountsAndAudits(t *testing.T) {
	svc, repo, _ := newService(t)
	_, err := svc.Login(t.Context(), "ana", "clave-mala")
	require.ErrorIs(t, err, domain.ErrInvalidCredentials)
	assert.Equal(t, int32(1), repo.user.FailedLogins)
	assert.Empty(t, repo.tokens)
	assert.Equal(t, []string{domain.AuditLoginFailed}, repo.actions())
}

func TestLogin_UnknownUser_RejectedAndAuditedWithoutUser(t *testing.T) {
	svc, repo, _ := newService(t)
	_, err := svc.Login(t.Context(), "nadie", testPassword)
	require.ErrorIs(t, err, domain.ErrInvalidCredentials)
	require.Len(t, repo.audits, 1)
	assert.Nil(t, repo.audits[0].UsuarioID)
}

func TestLogin_InactiveUser_Rejected(t *testing.T) {
	svc, repo, _ := newService(t)
	repo.user.Activo = false
	_, err := svc.Login(t.Context(), "ana", testPassword)
	require.ErrorIs(t, err, domain.ErrInvalidCredentials)
	assert.Equal(t, []string{domain.AuditLoginRejected}, repo.actions())
}

func TestLogin_FiveFailures_LocksEvenWithCorrectPasswordUntilExpiry(t *testing.T) {
	svc, repo, clock := newService(t)
	for range 5 {
		_, err := svc.Login(t.Context(), "ana", "clave-mala")
		require.ErrorIs(t, err, domain.ErrInvalidCredentials)
	}
	assert.Contains(t, repo.actions(), domain.AuditAccountLocked)
	lockedUntil := *repo.user.LockedUntil

	_, err := svc.Login(t.Context(), "ana", testPassword)
	require.ErrorIs(t, err, domain.ErrInvalidCredentials)
	assert.Equal(t, lockedUntil, *repo.user.LockedUntil, "attempts during a lock must not extend it")

	clock.t = clock.t.Add(domain.LockDuration)
	_, err = svc.Login(t.Context(), "ana", testPassword)
	require.NoError(t, err)
	assert.Nil(t, repo.user.LockedUntil)
	assert.Equal(t, int32(0), repo.user.FailedLogins)
}

func TestLogin_RepositoryError_IsNotCredentialsError(t *testing.T) {
	svc, repo, _ := newService(t)
	repo.failGet = errors.New("db down")
	_, err := svc.Login(t.Context(), "ana", testPassword)
	require.Error(t, err)
	assert.NotErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestLogin_CancelledWhileWaitingForHashSlot_ReturnsError(t *testing.T) {
	svc, _, _ := newService(t)
	for range maxConcurrentHashes {
		svc.hashSlots <- struct{}{}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := svc.Login(ctx, "ana", testPassword)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestLogin_IssuerError_FailsWithoutCredentialsError(t *testing.T) {
	svc, _, _ := newService(t)
	svc.tokens = fakeIssuer{err: errors.New("no key")}
	_, err := svc.Login(t.Context(), "ana", testPassword)
	require.Error(t, err)
	assert.NotErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestRefresh_ValidToken_RotatesWithinFamily(t *testing.T) {
	svc, repo, _ := newService(t)
	first, err := svc.Login(t.Context(), "ana", testPassword)
	require.NoError(t, err)

	second, err := svc.Refresh(t.Context(), first.RefreshToken)
	require.NoError(t, err)
	assert.NotEqual(t, first.RefreshToken, second.RefreshToken)
	require.Len(t, repo.tokens, 2)
	assert.NotNil(t, repo.tokens[0].RevokedAt)
	assert.Nil(t, repo.tokens[1].RevokedAt)
	assert.Equal(t, repo.tokens[0].FamilyID, repo.tokens[1].FamilyID)
}

func TestRefresh_ReusedToken_RevokesFamilyAndAudits(t *testing.T) {
	svc, repo, _ := newService(t)
	first, err := svc.Login(t.Context(), "ana", testPassword)
	require.NoError(t, err)
	second, err := svc.Refresh(t.Context(), first.RefreshToken)
	require.NoError(t, err)

	_, err = svc.Refresh(t.Context(), first.RefreshToken)
	require.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	assert.Contains(t, repo.actions(), domain.AuditRefreshReused)

	_, err = svc.Refresh(t.Context(), second.RefreshToken)
	assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken, "the legitimate descendant is revoked too")
}

func TestRefresh_InvalidTokens_Rejected(t *testing.T) {
	t.Run("unknown", func(t *testing.T) {
		svc, _, _ := newService(t)
		_, err := svc.Refresh(t.Context(), "nope")
		assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	})
	t.Run("expired", func(t *testing.T) {
		svc, _, clock := newService(t)
		s, err := svc.Login(t.Context(), "ana", testPassword)
		require.NoError(t, err)
		clock.t = clock.t.Add(auth.RefreshTokenTTL)
		_, err = svc.Refresh(t.Context(), s.RefreshToken)
		assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	})
	t.Run("inactive user revokes family", func(t *testing.T) {
		svc, repo, _ := newService(t)
		s, err := svc.Login(t.Context(), "ana", testPassword)
		require.NoError(t, err)
		repo.user.Activo = false
		_, err = svc.Refresh(t.Context(), s.RefreshToken)
		assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
		assert.NotNil(t, repo.tokens[0].RevokedAt)
	})
}

func TestLogout_OwnToken_RevokesFamilyWithoutLaterTheftAlert(t *testing.T) {
	svc, repo, _ := newService(t)
	s, err := svc.Login(t.Context(), "ana", testPassword)
	require.NoError(t, err)

	require.NoError(t, svc.Logout(t.Context(), domain.Principal{UserID: repo.user.ID}, s.RefreshToken))
	assert.NotNil(t, repo.tokens[0].RevokedAt)
	assert.Contains(t, repo.actions(), domain.AuditLogout)

	_, err = svc.Refresh(t.Context(), s.RefreshToken)
	require.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	assert.NotContains(t, repo.actions(), domain.AuditRefreshReused)
}

func TestLogout_OtherUsersOrUnknownToken_DoesNothing(t *testing.T) {
	svc, repo, _ := newService(t)
	s, err := svc.Login(t.Context(), "ana", testPassword)
	require.NoError(t, err)

	require.NoError(t, svc.Logout(t.Context(), domain.Principal{UserID: uuid.New()}, s.RefreshToken))
	require.NoError(t, svc.Logout(t.Context(), domain.Principal{UserID: repo.user.ID}, "nope"))
	assert.Nil(t, repo.tokens[0].RevokedAt)
	assert.NotContains(t, repo.actions(), domain.AuditLogout)
}
