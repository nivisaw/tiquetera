package postgres

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tiquetera/internal/auth"
	"tiquetera/internal/domain"
	"tiquetera/internal/service"
)

const usuarioTestPassword = "clave-segura-123"

type usuarioFixture struct {
	clienteFixture

	usuarios *service.UsuarioService
	auth     *service.AuthService
}

func newUsuarioFixture(t *testing.T) usuarioFixture {
	t.Helper()
	f := newClienteFixture(t)
	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	authSvc, err := service.NewAuthService(NewAuthRepository(f.pool), auth.NewTokenManager(key, f.clock), f.clock)
	require.NoError(t, err)
	return usuarioFixture{clienteFixture: f, usuarios: service.NewUsuarioService(NewUsuarioRepository(f.pool), f.clock), auth: authSvc}
}

func (f usuarioFixture) createUser(t *testing.T, username string, rol domain.Role) domain.User {
	t.Helper()
	user, err := f.usuarios.Create(t.Context(), f.usuarioID, domain.NewUser{
		Username: username, Password: usuarioTestPassword, Nombre: "Persona", Rol: rol,
	})
	require.NoError(t, err)
	return user
}

func newPtr[T any](v T) *T { return &v }

func TestUsuarioCreate_DuplicateCaseInsensitive_ReturnsConflict(t *testing.T) {
	f := newUsuarioFixture(t)
	created := f.createUser(t, "Caja1", domain.RoleCajero)
	assert.Equal(t, "Caja1", created.Username)

	_, err := f.usuarios.Create(t.Context(), f.usuarioID, domain.NewUser{
		Username: "caja1", Password: usuarioTestPassword, Nombre: "Otra", Rol: domain.RoleCajero,
	})
	require.ErrorIs(t, err, domain.ErrUsernameTaken)

	var audits int
	require.NoError(t, f.owner.QueryRow(t.Context(),
		`SELECT count(*) FROM audit_log WHERE accion = $1 AND entidad_id = $2`, domain.AuditUsuarioCreated, created.ID).Scan(&audits))
	assert.Equal(t, 1, audits)
}

func TestUsuarioList_PaginatesAndCounts(t *testing.T) {
	f := newUsuarioFixture(t)
	f.createUser(t, "caja1", domain.RoleCajero)
	f.createUser(t, "caja2", domain.RoleCajero)

	items, total, err := f.usuarios.List(t.Context(), 2, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total, "includes the fixture admin")
	require.Len(t, items, 1)
	assert.Equal(t, "caja2", items[0].Username)
}

func TestUsuarioUpdate_Deactivate_ClosesSessionsAndBlocksLogin(t *testing.T) {
	f := newUsuarioFixture(t)
	user := f.createUser(t, "caja1", domain.RoleCajero)
	session, err := f.auth.Login(t.Context(), "caja1", usuarioTestPassword)
	require.NoError(t, err)

	updated, err := f.usuarios.Update(t.Context(), f.usuarioID, user.ID, domain.UserPatch{Activo: newPtr(false)})
	require.NoError(t, err)
	assert.False(t, updated.Activo)

	_, err = f.auth.Refresh(t.Context(), session.RefreshToken)
	require.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	_, err = f.auth.Login(t.Context(), "caja1", usuarioTestPassword)
	require.ErrorIs(t, err, domain.ErrInvalidCredentials)
}

func TestCurrentPrincipal_AfterDeactivateOrRoleChange_AppliesImmediately(t *testing.T) {
	f := newUsuarioFixture(t)
	user := f.createUser(t, "caja1", domain.RoleCajero)
	token := domain.Principal{UserID: user.ID, Rol: domain.RoleCajero}

	_, err := f.usuarios.Update(t.Context(), f.usuarioID, user.ID, domain.UserPatch{Rol: newPtr(domain.RoleAdmin)})
	require.NoError(t, err)
	got, err := f.auth.CurrentPrincipal(t.Context(), token)
	require.NoError(t, err)
	assert.Equal(t, domain.RoleAdmin, got.Rol)

	_, err = f.usuarios.Update(t.Context(), f.usuarioID, user.ID, domain.UserPatch{Activo: newPtr(false)})
	require.NoError(t, err)
	_, err = f.auth.CurrentPrincipal(t.Context(), token)
	require.ErrorIs(t, err, domain.ErrInvalidAccessToken)
	_, err = f.auth.CurrentPrincipal(t.Context(), domain.Principal{UserID: uuid.New()})
	require.ErrorIs(t, err, domain.ErrInvalidAccessToken)
}

func TestUsuarioUpdate_LastAdmin_Rejected(t *testing.T) {
	f := newUsuarioFixture(t)
	_, err := f.usuarios.Update(t.Context(), f.usuarioID, f.usuarioID, domain.UserPatch{Activo: newPtr(false)})
	require.ErrorIs(t, err, domain.ErrLastAdmin)
	_, err = f.usuarios.Update(t.Context(), f.usuarioID, f.usuarioID, domain.UserPatch{Rol: newPtr(domain.RoleCajero)})
	require.ErrorIs(t, err, domain.ErrLastAdmin)

	got, err := f.usuarios.Get(t.Context(), f.usuarioID)
	require.NoError(t, err)
	assert.True(t, got.IsActiveAdmin())
}

func TestUsuarioUpdate_AdminsDeactivatingEachOther_LeavesOneActive(t *testing.T) {
	f := newUsuarioFixture(t)
	other := f.createUser(t, "admin2", domain.RoleAdmin)

	targets := []uuid.UUID{f.usuarioID, other.ID}
	errs := make([]error, len(targets))
	var wg sync.WaitGroup
	for i, target := range targets {
		wg.Go(func() {
			_, errs[i] = f.usuarios.Update(t.Context(), targets[1-i], target, domain.UserPatch{Activo: newPtr(false)})
		})
	}
	wg.Wait()

	assert.Equal(t, 1, countActiveAdmins(t, f.pool))
	assert.True(t, errors.Is(errs[0], domain.ErrLastAdmin) != errors.Is(errs[1], domain.ErrLastAdmin), "%v / %v", errs[0], errs[1])
}

func countActiveAdmins(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()
	var n int
	require.NoError(t, pool.QueryRow(t.Context(), `SELECT count(*) FROM usuarios WHERE rol = 'admin' AND activo`).Scan(&n))
	return n
}

func TestUsuarioResetPassword_ReplacesPasswordClearsLockAndClosesSessions(t *testing.T) {
	f := newUsuarioFixture(t)
	user := f.createUser(t, "caja1", domain.RoleCajero)
	session, err := f.auth.Login(t.Context(), "caja1", usuarioTestPassword)
	require.NoError(t, err)
	for range domain.MaxFailedLogins {
		_, err = f.auth.Login(t.Context(), "caja1", "clave-incorrecta")
		require.ErrorIs(t, err, domain.ErrInvalidCredentials)
	}

	const replacement = "otra-clave-segura-456"
	require.NoError(t, f.usuarios.ResetPassword(t.Context(), f.usuarioID, user.ID, replacement))

	_, err = f.auth.Refresh(t.Context(), session.RefreshToken)
	require.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
	_, err = f.auth.Login(t.Context(), "caja1", usuarioTestPassword)
	require.ErrorIs(t, err, domain.ErrInvalidCredentials, "old password no longer works")
	_, err = f.auth.Login(t.Context(), "caja1", replacement)
	require.NoError(t, err, "lock was cleared")

	require.ErrorIs(t, f.usuarios.ResetPassword(t.Context(), f.usuarioID, uuid.New(), replacement), domain.ErrNotFound)
}

func TestUsuarioResetPassword_ConcurrentRefresh_NoDeadlockNoSurvivingSession(t *testing.T) {
	f := newUsuarioFixture(t)
	user := f.createUser(t, "caja1", domain.RoleCajero)

	for range 10 {
		session, err := f.auth.Login(t.Context(), "caja1", usuarioTestPassword)
		require.NoError(t, err)

		var refreshErr, resetErr error
		var wg sync.WaitGroup
		wg.Go(func() { _, refreshErr = f.auth.Refresh(t.Context(), session.RefreshToken) })
		wg.Go(func() { resetErr = f.usuarios.ResetPassword(t.Context(), f.usuarioID, user.ID, usuarioTestPassword) })
		wg.Wait()

		require.NoError(t, resetErr)
		if refreshErr != nil {
			require.ErrorIs(t, refreshErr, domain.ErrInvalidRefreshToken)
		}
		var active int
		require.NoError(t, f.pool.QueryRow(t.Context(),
			`SELECT count(*) FROM refresh_tokens WHERE usuario_id = $1 AND revoked_at IS NULL`, user.ID).Scan(&active))
		assert.Zero(t, active)
	}
}

func TestClienteListConsumos_NewestFirstPaginated(t *testing.T) {
	f := newUsuarioFixture(t)
	created := f.create(t, clienteTestPhone)
	consumos := service.NewConsumoService(NewConsumoRepository(f.pool), f.clock)
	for range 3 {
		_, err := consumos.Consume(t.Context(), f.usuarioID, uuid.New(), created.QRPayload)
		require.NoError(t, err)
	}

	items, total, err := f.svc.ListConsumos(t.Context(), created.Cliente.ID, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, items, 2)
	assert.Equal(t, int32(17), items[0].SaldoResultante)
	assert.Equal(t, int32(18), items[1].SaldoResultante)
	assert.Equal(t, f.usuarioID, items[0].CajeroID)

	items, _, err = f.svc.ListConsumos(t.Context(), created.Cliente.ID, 2, 2)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, int32(19), items[0].SaldoResultante)

	_, _, err = f.svc.ListConsumos(t.Context(), uuid.New(), 1, 2)
	require.ErrorIs(t, err, domain.ErrNotFound)
}
