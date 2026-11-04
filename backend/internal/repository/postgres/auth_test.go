package postgres

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tiquetera/internal/auth"
	"tiquetera/internal/domain"
	"tiquetera/internal/platform"
	"tiquetera/internal/service"
)

const authTestPassword = "clave-correcta-123"

type movableClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *movableClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *movableClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

type authFixture struct {
	svc    *service.AuthService
	owner  *pgx.Conn
	clock  *movableClock
	userID uuid.UUID
}

// newAuthFixture runs the service against a real database as the app role.
func newAuthFixture(t *testing.T) authFixture {
	t.Helper()
	ctx, ownerURL, appURL := startDB(t)
	require.NoError(t, newMigrator(t, ownerURL).Up())
	owner := connect(ctx, t, ownerURL)
	_, err := owner.Exec(ctx, fmt.Sprintf(`ALTER ROLE tiquetera_app LOGIN PASSWORD '%s'`, appPassword))
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, appURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	hash, err := platform.HashPassword(authTestPassword)
	require.NoError(t, err)
	var userID uuid.UUID
	require.NoError(t, owner.QueryRow(ctx,
		`INSERT INTO usuarios (username, password_hash, nombre, rol) VALUES ('Ana', $1, 'Ana', 'cajero') RETURNING id`, hash,
	).Scan(&userID))

	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	clock := &movableClock{t: time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)}
	svc, err := service.NewAuthService(NewAuthRepository(pool), auth.NewTokenManager(key, clock), clock)
	require.NoError(t, err)
	return authFixture{svc: svc, owner: owner, clock: clock, userID: userID}
}

func (f authFixture) countAudit(t *testing.T, action string) int {
	t.Helper()
	var n int
	require.NoError(t, f.owner.QueryRow(t.Context(), `SELECT count(*) FROM audit_log WHERE accion = $1`, action).Scan(&n))
	return n
}

func TestAuthFlow_LoginRotateReuse_RevokesFamilyAndPersistsAudit(t *testing.T) {
	f := newAuthFixture(t)
	ctx := t.Context()

	first, err := f.svc.Login(ctx, "ana", authTestPassword)
	require.NoError(t, err, "username is case-insensitive")
	assert.Equal(t, 1, f.countAudit(t, domain.AuditLoginOK))

	second, err := f.svc.Refresh(ctx, first.RefreshToken)
	require.NoError(t, err)

	_, err = f.svc.Refresh(ctx, first.RefreshToken)
	require.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	assert.Equal(t, 1, f.countAudit(t, domain.AuditRefreshReused), "revocation must commit despite the failure")

	_, err = f.svc.Refresh(ctx, second.RefreshToken)
	require.ErrorIs(t, err, domain.ErrInvalidRefreshToken)

	var active int
	require.NoError(t, f.owner.QueryRow(ctx, `SELECT count(*) FROM refresh_tokens WHERE revoked_at IS NULL`).Scan(&active))
	assert.Zero(t, active)
}

func TestAuthFlow_ConcurrentRefresh_ExactlyOneRotationWins(t *testing.T) {
	f := newAuthFixture(t)
	s, err := f.svc.Login(t.Context(), "ana", authTestPassword)
	require.NoError(t, err)

	const callers = 8
	results := make([]error, callers)
	var wg sync.WaitGroup
	for i := range callers {
		wg.Go(func() {
			_, results[i] = f.svc.Refresh(t.Context(), s.RefreshToken)
		})
	}
	wg.Wait()

	wins := 0
	for _, err := range results {
		if err == nil {
			wins++
			continue
		}
		require.True(t, errors.Is(err, domain.ErrInvalidRefreshToken), "unexpected error: %v", err)
	}
	assert.Equal(t, 1, wins)
}

func TestAuthFlow_FiveFailures_PersistLockUntilExpiry(t *testing.T) {
	f := newAuthFixture(t)
	ctx := t.Context()
	for range domain.MaxFailedLogins {
		_, err := f.svc.Login(ctx, "ana", "clave-mala")
		require.ErrorIs(t, err, domain.ErrInvalidCredentials)
	}
	assert.Equal(t, 5, f.countAudit(t, domain.AuditLoginFailed))
	assert.Equal(t, 1, f.countAudit(t, domain.AuditAccountLocked))

	_, err := f.svc.Login(ctx, "ana", authTestPassword)
	require.ErrorIs(t, err, domain.ErrInvalidCredentials)

	f.clock.advance(domain.LockDuration + time.Second)
	_, err = f.svc.Login(ctx, "ana", authTestPassword)
	require.NoError(t, err)

	var failed int
	var locked *time.Time
	require.NoError(t, f.owner.QueryRow(ctx, `SELECT failed_logins, locked_until FROM usuarios WHERE id = $1`, f.userID).Scan(&failed, &locked))
	assert.Zero(t, failed)
	assert.Nil(t, locked)
}

func TestAuthFlow_UnknownUser_AuditsWithNullUser(t *testing.T) {
	f := newAuthFixture(t)
	_, err := f.svc.Login(t.Context(), "nadie", "x")
	require.ErrorIs(t, err, domain.ErrInvalidCredentials)

	var withNull int
	require.NoError(t, f.owner.QueryRow(t.Context(),
		`SELECT count(*) FROM audit_log WHERE accion = $1 AND usuario_id IS NULL AND datos = '{}'`, domain.AuditLoginFailed,
	).Scan(&withNull))
	assert.Equal(t, 1, withNull)
}

func TestAuthFlow_Logout_RevokesFamily(t *testing.T) {
	f := newAuthFixture(t)
	s, err := f.svc.Login(t.Context(), "ana", authTestPassword)
	require.NoError(t, err)

	require.NoError(t, f.svc.Logout(t.Context(), domain.Principal{UserID: f.userID, Rol: domain.RoleCajero}, s.RefreshToken))
	assert.Equal(t, 1, f.countAudit(t, domain.AuditLogout))

	_, err = f.svc.Refresh(t.Context(), s.RefreshToken)
	assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	assert.Zero(t, f.countAudit(t, domain.AuditRefreshReused))
}
