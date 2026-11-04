package http

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tiquetera/internal/auth"
	"tiquetera/internal/domain"
)

const (
	validUsuarioReq = `{"username":"caja1","password":"clave-secreta-123","nombre":"Caja","rol":"cajero"}`
	validResetReq   = `{"password":"clave-secreta-123"}`
)

type fakeUsuarios struct {
	err      error
	user     domain.User
	patch    domain.UserPatch
	input    domain.NewUser
	password string
	page     int32
	size     int32
	actor    uuid.UUID
	calls    int
}

func (f *fakeUsuarios) Create(_ context.Context, actor uuid.UUID, input domain.NewUser) (domain.User, error) {
	f.calls++
	f.actor, f.input = actor, input
	return f.user, f.err
}

func (f *fakeUsuarios) Get(context.Context, uuid.UUID) (domain.User, error) {
	f.calls++
	return f.user, f.err
}

func (f *fakeUsuarios) List(_ context.Context, page, size int32) ([]domain.User, int64, error) {
	f.calls++
	f.page, f.size = page, size
	return []domain.User{f.user}, 1, f.err
}

func (f *fakeUsuarios) Update(_ context.Context, _, _ uuid.UUID, patch domain.UserPatch) (domain.User, error) {
	f.calls++
	f.patch = patch
	return f.user, f.err
}

func (f *fakeUsuarios) ResetPassword(_ context.Context, _, _ uuid.UUID, password string) error {
	f.calls++
	f.password = password
	return f.err
}

type usuariosEnv struct {
	router http.Handler
	svc    *fakeUsuarios
	tokens *auth.TokenManager
	userID uuid.UUID
}

func newUsuariosEnv(t *testing.T) usuariosEnv {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	tokens := auth.NewTokenManager(key, &stubClock{t: time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)})
	svc := &fakeUsuarios{user: domain.User{
		ID: uuid.New(), Username: "caja1", Nombre: "Caja", Rol: domain.RoleCajero, Activo: true, PasswordHash: "hash-secreto",
	}}
	router := NewRouter(Deps{
		Log: slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)), DB: fakePinger{},
		Tokens: tokens, Auth: &fakeAuth{}, Usuarios: svc,
	})
	return usuariosEnv{router: router, svc: svc, tokens: tokens, userID: uuid.New()}
}

func (e usuariosEnv) call(t *testing.T, rol domain.Role, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	token, err := e.tokens.IssueAccess(domain.Principal{UserID: e.userID, Rol: rol})
	require.NoError(t, err)
	req := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	return rec
}

func usuarioRoutes() map[string][2]string {
	item := "/v1/usuarios/" + uuid.NewString()
	return map[string][2]string{
		"create": {http.MethodPost, "/v1/usuarios"},
		"list":   {http.MethodGet, "/v1/usuarios"},
		"get":    {http.MethodGet, item},
		"update": {http.MethodPatch, item},
		"reset":  {http.MethodPut, item + "/password"},
	}
}

func TestUsuarios_NonAdminOrAnonymous_Rejected(t *testing.T) {
	e := newUsuariosEnv(t)
	for name, route := range usuarioRoutes() {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, http.StatusForbidden, e.call(t, domain.RoleCajero, route[0], route[1], validUsuarioReq).Code)

			req := httptest.NewRequestWithContext(t.Context(), route[0], route[1], strings.NewReader(validUsuarioReq))
			rec := httptest.NewRecorder()
			e.router.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	}
	assert.Zero(t, e.svc.calls)
}

func TestUsuarioCreate_Valid_Returns201WithoutHash(t *testing.T) {
	e := newUsuariosEnv(t)
	rec := e.call(t, domain.RoleAdmin, http.MethodPost, "/v1/usuarios", validUsuarioReq)
	require.Equal(t, http.StatusCreated, rec.Code)
	assert.JSONEq(t, fmt.Sprintf(`{"id":%q,"username":"caja1","nombre":"Caja","rol":"cajero","activo":true}`, e.svc.user.ID), rec.Body.String())
	assert.Equal(t, e.userID, e.svc.actor)
	assert.Equal(t, domain.RoleCajero, e.svc.input.Rol)
}

func TestUsuarioList_Pagination(t *testing.T) {
	e := newUsuariosEnv(t)
	rec := e.call(t, domain.RoleAdmin, http.MethodGet, "/v1/usuarios?page=2&size=50", "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, int32(2), e.svc.page)
	assert.Equal(t, int32(50), e.svc.size)
	assert.NotContains(t, rec.Body.String(), "hash-secreto")

	for _, query := range []string{"page=0", "size=51", "size=x"} {
		assert.Equal(t, http.StatusBadRequest, e.call(t, domain.RoleAdmin, http.MethodGet, "/v1/usuarios?"+query, "").Code, query)
	}
}

func TestUsuarioGet_MalformedID_Returns404(t *testing.T) {
	e := newUsuariosEnv(t)
	assert.Equal(t, http.StatusOK, e.call(t, domain.RoleAdmin, http.MethodGet, "/v1/usuarios/"+uuid.NewString(), "").Code)
	assert.Equal(t, http.StatusNotFound, e.call(t, domain.RoleAdmin, http.MethodGet, "/v1/usuarios/x", "").Code)
	assert.Equal(t, http.StatusNotFound, e.call(t, domain.RoleAdmin, http.MethodPatch, "/v1/usuarios/x", `{"activo":false}`).Code)
	assert.Equal(t, http.StatusNotFound, e.call(t, domain.RoleAdmin, http.MethodPut, "/v1/usuarios/x/password", validResetReq).Code)
}

func TestUsuarioUpdate_Valid_PassesPatch(t *testing.T) {
	e := newUsuariosEnv(t)
	rec := e.call(t, domain.RoleAdmin, http.MethodPatch, "/v1/usuarios/"+uuid.NewString(), `{"activo":false}`)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, e.svc.patch.Activo)
	assert.False(t, *e.svc.patch.Activo)
	assert.Nil(t, e.svc.patch.Rol)
}

func TestUsuarioWrites_InvalidBody_Returns400WithoutCallingService(t *testing.T) {
	item := "/v1/usuarios/" + uuid.NewString()
	cases := map[string][3]string{
		"create not json":      {http.MethodPost, "/v1/usuarios", `{`},
		"create unknown field": {http.MethodPost, "/v1/usuarios", `{"username":"a","password_hash":"x"}`},
		"patch empty":          {http.MethodPatch, item, `{}`},
		"patch null nombre":    {http.MethodPatch, item, `{"nombre":null}`},
		"patch null rol":       {http.MethodPatch, item, `{"rol":null}`},
		"patch null activo":    {http.MethodPatch, item, `{"activo":null}`},
		"patch password":       {http.MethodPatch, item, `{"password":"clave-secreta-123"}`},
		"patch not json":       {http.MethodPatch, item, `[`},
		"reset not json":       {http.MethodPut, item + "/password", `{`},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := newUsuariosEnv(t)
			assert.Equal(t, http.StatusBadRequest, e.call(t, domain.RoleAdmin, tc[0], tc[1], tc[2]).Code)
			assert.Zero(t, e.svc.calls)
		})
	}
}

func TestUsuarioResetPassword_Valid_Returns204(t *testing.T) {
	e := newUsuariosEnv(t)
	rec := e.call(t, domain.RoleAdmin, http.MethodPut, "/v1/usuarios/"+uuid.NewString()+"/password", validResetReq)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "clave-secreta-123", e.svc.password)
}

func TestUsuarioErrors_MapToProblem(t *testing.T) {
	cases := map[string]struct {
		err    error
		status int
		code   string
	}{
		"duplicate":  {domain.ErrUsernameTaken, http.StatusConflict, "USERNAME_DUPLICADO"},
		"last admin": {domain.ErrLastAdmin, http.StatusUnprocessableEntity, "ULTIMO_ADMIN"},
		"validation": {domain.ValidationError{Message: "x"}, http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		"not found":  {domain.ErrNotFound, http.StatusNotFound, "NOT_FOUND"},
	}
	for name, tc := range cases {
		for routeName, route := range usuarioRoutes() {
			body := map[string]string{"create": validUsuarioReq, "update": `{"activo":false}`, "reset": validResetReq}[routeName]
			t.Run(name+"/"+routeName, func(t *testing.T) {
				e := newUsuariosEnv(t)
				e.svc.err = tc.err
				rec := e.call(t, domain.RoleAdmin, route[0], route[1], body)
				assert.Equal(t, tc.status, rec.Code)
				assert.Contains(t, rec.Body.String(), tc.code)
			})
		}
	}
}

func TestClienteConsumos_Admin_ReturnsPage(t *testing.T) {
	e := newClientesEnv(t)
	rec := e.call(t, domain.RoleAdmin, http.MethodGet, "/v1/clientes/"+uuid.NewString()+"/consumos?page=2&size=10", "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"saldo_resultante":4`)
	assert.Contains(t, rec.Body.String(), `"total":1`)
	assert.Equal(t, int32(2), e.svc.page)
	assert.Equal(t, int32(10), e.svc.size)
}

func TestClienteConsumos_Errors(t *testing.T) {
	e := newClientesEnv(t)
	path := "/v1/clientes/" + uuid.NewString() + "/consumos"
	assert.Equal(t, http.StatusForbidden, e.call(t, domain.RoleCajero, http.MethodGet, path, "").Code)
	assert.Equal(t, http.StatusBadRequest, e.call(t, domain.RoleAdmin, http.MethodGet, path+"?size=51", "").Code)
	assert.Equal(t, http.StatusNotFound, e.call(t, domain.RoleAdmin, http.MethodGet, "/v1/clientes/x/consumos", "").Code)

	e.svc.err = domain.ErrNotFound
	assert.Equal(t, http.StatusNotFound, e.call(t, domain.RoleAdmin, http.MethodGet, path, "").Code)
}
