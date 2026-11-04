package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tiquetera/internal/domain"
	"tiquetera/internal/platform"
)

const usuarioTestPassword = "clave-segura-123"

type fakeUsuarioRepo struct {
	user         domain.User
	getErr       error
	createErr    error
	activeAdmins int
	failOn       string
	calls        []string
	audits       []domain.AuditEntry
	passwordHash string
	limit        int32
	offset       int32
}

func (f *fakeUsuarioRepo) step(name string) error {
	f.calls = append(f.calls, name)
	if f.failOn == name {
		return errRepo
	}
	return nil
}

func (f *fakeUsuarioRepo) InTx(_ context.Context, fn func(UsuarioRepository) error) error {
	return fn(f)
}

func (f *fakeUsuarioRepo) CreateUser(_ context.Context, u domain.NewUser, hash string) (domain.User, error) {
	f.passwordHash = hash
	return domain.User{ID: uuid.New(), Username: u.Username, Nombre: u.Nombre, Rol: u.Rol, Activo: true}, f.createErr
}

func (f *fakeUsuarioRepo) GetUser(context.Context, uuid.UUID) (domain.User, error) {
	return f.user, f.getErr
}

func (f *fakeUsuarioRepo) GetUserForUpdate(context.Context, uuid.UUID) (domain.User, error) {
	return f.user, f.getErr
}

func (f *fakeUsuarioRepo) ListUsers(_ context.Context, limit, offset int32) ([]domain.User, int64, error) {
	f.limit, f.offset = limit, offset
	return []domain.User{f.user}, 7, f.step("ListUsers")
}

func (f *fakeUsuarioRepo) LockActiveAdmins(context.Context) (int, error) {
	return f.activeAdmins, f.step("LockActiveAdmins")
}

func (f *fakeUsuarioRepo) UpdateUser(_ context.Context, u domain.User) (domain.User, error) {
	return u, f.step("UpdateUser")
}

func (f *fakeUsuarioRepo) SetPassword(_ context.Context, _ uuid.UUID, hash string) error {
	f.passwordHash = hash
	return f.step("SetPassword")
}

func (f *fakeUsuarioRepo) RevokeUserSessions(context.Context, uuid.UUID, time.Time) error {
	return f.step("RevokeUserSessions")
}

func (f *fakeUsuarioRepo) AddAudit(_ context.Context, e domain.AuditEntry) error {
	f.audits = append(f.audits, e)
	return f.step("AddAudit")
}

func newUsuarioSvc(repo *fakeUsuarioRepo) *UsuarioService {
	return NewUsuarioService(repo, platform.SystemClock{})
}

func ptr[T any](v T) *T { return &v }

func activeUser(rol domain.Role) domain.User {
	return domain.User{ID: uuid.New(), Username: "ana", Nombre: "Ana", Rol: rol, Activo: true}
}

func TestUsuarioCreate_Valid_StoresHashAndAudits(t *testing.T) {
	repo := &fakeUsuarioRepo{}
	actor := uuid.New()
	created, err := newUsuarioSvc(repo).Create(t.Context(), actor, domain.NewUser{
		Username: "caja1", Password: usuarioTestPassword, Nombre: "Caja", Rol: domain.RoleCajero,
	})
	require.NoError(t, err)
	assert.Equal(t, "caja1", created.Username)

	match, err := platform.VerifyPassword(usuarioTestPassword, repo.passwordHash)
	require.NoError(t, err)
	assert.True(t, match)
	require.Len(t, repo.audits, 1)
	assert.Equal(t, domain.AuditUsuarioCreated, repo.audits[0].Accion)
	assert.Equal(t, actor, *repo.audits[0].UsuarioID)
}

func TestUsuarioCreate_Invalid_DoesNotTouchRepo(t *testing.T) {
	repo := &fakeUsuarioRepo{}
	_, err := newUsuarioSvc(repo).Create(t.Context(), uuid.New(), domain.NewUser{Username: "x"})
	var invalid domain.ValidationError
	require.ErrorAs(t, err, &invalid)
	assert.Empty(t, repo.passwordHash)
}

func TestUsuarioCreate_UsernameTaken_ReturnsDomainError(t *testing.T) {
	repo := &fakeUsuarioRepo{createErr: domain.ErrUsernameTaken}
	_, err := newUsuarioSvc(repo).Create(t.Context(), uuid.New(), domain.NewUser{
		Username: "caja1", Password: usuarioTestPassword, Nombre: "Caja", Rol: domain.RoleCajero,
	})
	require.ErrorIs(t, err, domain.ErrUsernameTaken)
	assert.Empty(t, repo.audits)
}

func TestUsuarioCreate_CanceledWhileWaitingHashSlot_ReturnsError(t *testing.T) {
	svc := newUsuarioSvc(&fakeUsuarioRepo{})
	for range maxConcurrentHashes {
		svc.hashSlots <- struct{}{}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := svc.Create(ctx, uuid.New(), domain.NewUser{
		Username: "caja1", Password: usuarioTestPassword, Nombre: "Caja", Rol: domain.RoleCajero,
	})
	require.ErrorIs(t, err, context.Canceled)
}

func TestUsuarioGetAndList_PassThrough(t *testing.T) {
	repo := &fakeUsuarioRepo{user: activeUser(domain.RoleCajero)}
	svc := newUsuarioSvc(repo)

	got, err := svc.Get(t.Context(), repo.user.ID)
	require.NoError(t, err)
	assert.Equal(t, repo.user, got)

	_, total, err := svc.List(t.Context(), 3, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(7), total)
	assert.Equal(t, int32(20), repo.limit)
	assert.Equal(t, int32(40), repo.offset)

	repo.getErr = domain.ErrNotFound
	_, err = svc.Get(t.Context(), repo.user.ID)
	require.ErrorIs(t, err, domain.ErrNotFound)

	repo.failOn = "ListUsers"
	_, _, err = svc.List(t.Context(), 1, 20)
	require.ErrorIs(t, err, errRepo)
}

func TestUsuarioUpdate_LastAdminLosingAccess_Rejected(t *testing.T) {
	cases := map[string]domain.UserPatch{
		"deactivate": {Activo: ptr(false)},
		"demote":     {Rol: ptr(domain.RoleCajero)},
	}
	for name, patch := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &fakeUsuarioRepo{user: activeUser(domain.RoleAdmin), activeAdmins: 1}
			_, err := newUsuarioSvc(repo).Update(t.Context(), uuid.New(), repo.user.ID, patch)
			require.ErrorIs(t, err, domain.ErrLastAdmin)
			assert.Equal(t, []string{"LockActiveAdmins"}, repo.calls)
		})
	}
}

func TestUsuarioUpdate_SessionRevocation(t *testing.T) {
	cases := map[string]struct {
		user   domain.User
		patch  domain.UserPatch
		admins int
		revoke bool
	}{
		"admin deactivated with another left": {activeUser(domain.RoleAdmin), domain.UserPatch{Activo: ptr(false)}, 2, true},
		"cashier deactivated":                 {activeUser(domain.RoleCajero), domain.UserPatch{Activo: ptr(false)}, 1, true},
		"cashier promoted":                    {activeUser(domain.RoleCajero), domain.UserPatch{Rol: ptr(domain.RoleAdmin)}, 1, true},
		"same role sent":                      {activeUser(domain.RoleAdmin), domain.UserPatch{Rol: ptr(domain.RoleAdmin)}, 1, false},
		"last admin stays active":             {activeUser(domain.RoleAdmin), domain.UserPatch{Activo: ptr(true)}, 1, false},
		"rename only":                         {activeUser(domain.RoleAdmin), domain.UserPatch{Nombre: ptr("Rosa")}, 0, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &fakeUsuarioRepo{user: tc.user, activeAdmins: tc.admins}
			_, err := newUsuarioSvc(repo).Update(t.Context(), uuid.New(), tc.user.ID, tc.patch)
			require.NoError(t, err)
			if tc.revoke {
				assert.Contains(t, repo.calls, "RevokeUserSessions")
			} else {
				assert.NotContains(t, repo.calls, "RevokeUserSessions")
			}
			assert.Contains(t, repo.calls, "AddAudit")
		})
	}
}

func TestUsuarioUpdate_RenameOnly_DoesNotLockAdmins(t *testing.T) {
	repo := &fakeUsuarioRepo{user: activeUser(domain.RoleAdmin)}
	updated, err := newUsuarioSvc(repo).Update(t.Context(), uuid.New(), repo.user.ID, domain.UserPatch{Nombre: ptr(" Rosa ")})
	require.NoError(t, err)
	assert.Equal(t, "Rosa", updated.Nombre)
	assert.NotContains(t, repo.calls, "LockActiveAdmins")
}

func TestUsuarioUpdate_Errors_Propagate(t *testing.T) {
	var invalid domain.ValidationError
	_, err := newUsuarioSvc(&fakeUsuarioRepo{}).Update(t.Context(), uuid.New(), uuid.New(), domain.UserPatch{Nombre: ptr("A")})
	require.ErrorAs(t, err, &invalid)

	notFound := &fakeUsuarioRepo{getErr: domain.ErrNotFound}
	_, err = newUsuarioSvc(notFound).Update(t.Context(), uuid.New(), uuid.New(), domain.UserPatch{Nombre: ptr("Rosa")})
	require.ErrorIs(t, err, domain.ErrNotFound)

	for _, step := range []string{"LockActiveAdmins", "UpdateUser", "RevokeUserSessions", "AddAudit"} {
		repo := &fakeUsuarioRepo{user: activeUser(domain.RoleCajero), failOn: step}
		_, err = newUsuarioSvc(repo).Update(t.Context(), uuid.New(), repo.user.ID, domain.UserPatch{Activo: ptr(false)})
		require.ErrorIs(t, err, errRepo, step)
	}
}

func TestUsuarioResetPassword_Valid_RevokesAndAudits(t *testing.T) {
	repo := &fakeUsuarioRepo{user: activeUser(domain.RoleCajero)}
	require.NoError(t, newUsuarioSvc(repo).ResetPassword(t.Context(), uuid.New(), repo.user.ID, usuarioTestPassword))

	assert.Equal(t, []string{"SetPassword", "RevokeUserSessions", "AddAudit"}, repo.calls)
	assert.Equal(t, domain.AuditPasswordChanged, repo.audits[0].Accion)
	assert.NotContains(t, repo.passwordHash, usuarioTestPassword)
}

func TestUsuarioResetPassword_Errors(t *testing.T) {
	var invalid domain.ValidationError
	err := newUsuarioSvc(&fakeUsuarioRepo{}).ResetPassword(t.Context(), uuid.New(), uuid.New(), "corta")
	require.ErrorAs(t, err, &invalid)

	notFound := &fakeUsuarioRepo{getErr: domain.ErrNotFound}
	err = newUsuarioSvc(notFound).ResetPassword(t.Context(), uuid.New(), uuid.New(), usuarioTestPassword)
	require.ErrorIs(t, err, domain.ErrNotFound)

	for _, step := range []string{"SetPassword", "RevokeUserSessions", "AddAudit"} {
		repo := &fakeUsuarioRepo{failOn: step}
		err = newUsuarioSvc(repo).ResetPassword(t.Context(), uuid.New(), uuid.New(), usuarioTestPassword)
		require.ErrorIs(t, err, errRepo, step)
	}
}

func TestClienteListConsumos_ComputesOffsetAndChecksClient(t *testing.T) {
	repo := &fakeClienteRepo{}
	svc := NewClienteService(repo, platform.SystemClock{})
	items, total, err := svc.ListConsumos(t.Context(), uuid.New(), 2, 10)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, int64(7), total)
	assert.Equal(t, int32(10), repo.limit)
	assert.Equal(t, int32(10), repo.offset)

	repo.failOn = "ListConsumos"
	_, _, err = svc.ListConsumos(t.Context(), uuid.New(), 1, 10)
	require.ErrorIs(t, err, errRepo)

	repo = &fakeClienteRepo{getErr: domain.ErrNotFound}
	_, _, err = NewClienteService(repo, platform.SystemClock{}).ListConsumos(t.Context(), uuid.New(), 1, 10)
	require.ErrorIs(t, err, domain.ErrNotFound)
	assert.NotContains(t, repo.calls, "ListConsumos")
}
