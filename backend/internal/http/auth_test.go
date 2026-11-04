package http

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
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

const validLoginBody = `{"username":"ana","password":"clave-secreta-123"}`

type stubClock struct{ t time.Time }

func (c *stubClock) Now() time.Time { return c.t }

type fakeAuth struct {
	session    domain.Session
	err        error
	loginCalls int
	logoutUser uuid.UUID
	currentErr error
	currentRol domain.Role
}

func (f *fakeAuth) CurrentPrincipal(_ context.Context, p domain.Principal) (domain.Principal, error) {
	if f.currentRol != "" {
		p.Rol = f.currentRol
	}
	return p, f.currentErr
}

func TestAuthenticate_UsesCurrentUserState(t *testing.T) {
	cases := map[string]struct {
		err    error
		rol    domain.Role
		status int
	}{
		"inactive user":    {err: domain.ErrInvalidAccessToken, status: http.StatusUnauthorized},
		"database down":    {err: errors.New("db down"), status: http.StatusInternalServerError},
		"demoted admin":    {rol: domain.RoleCajero, status: http.StatusForbidden},
		"still authorized": {status: http.StatusNotFound},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := newAuthEnv(t, false)
			e.svc.currentErr, e.svc.currentRol = tc.err, tc.rol
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/clientes/no-es-uuid/consumos", nil)
			req.Header.Set("Authorization", "Bearer "+e.bearer(t, domain.RoleAdmin))
			rec := httptest.NewRecorder()
			e.router.ServeHTTP(rec, req)
			assert.Equal(t, tc.status, rec.Code)
		})
	}
}

func (f *fakeAuth) Login(context.Context, string, string) (domain.Session, error) {
	f.loginCalls++
	return f.session, f.err
}

func (f *fakeAuth) Refresh(context.Context, string) (domain.Session, error) {
	return f.session, f.err
}

func (f *fakeAuth) Logout(_ context.Context, p domain.Principal, _ string) error {
	f.logoutUser = p.UserID
	return f.err
}

type authEnv struct {
	router http.Handler
	svc    *fakeAuth
	tokens *auth.TokenManager
	clock  *stubClock
	logs   *bytes.Buffer
}

func newAuthEnv(t *testing.T, trustProxy bool) authEnv {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	clock := &stubClock{t: time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)}
	tokens := auth.NewTokenManager(key, clock)
	svc := &fakeAuth{session: domain.Session{
		AccessToken: "a", RefreshToken: "r", ExpiresIn: 900,
		User: domain.User{ID: uuid.New(), Nombre: "Ana", Rol: domain.RoleCajero, PasswordHash: "hash-secreto"},
	}}
	logs := &bytes.Buffer{}
	router := NewRouter(Deps{
		Log: slog.New(slog.NewJSONHandler(logs, nil)), DB: fakePinger{},
		Auth: svc, Tokens: tokens, TrustProxyHeaders: trustProxy,
	})
	return authEnv{router: router, svc: svc, tokens: tokens, clock: clock, logs: logs}
}

func (e authEnv) bearer(t *testing.T, rol domain.Role) string {
	t.Helper()
	token, err := e.tokens.IssueAccess(domain.Principal{UserID: uuid.New(), Rol: rol})
	require.NoError(t, err)
	return token
}

func post(h http.Handler, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, path, strings.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestLoginHandler_Valid_ReturnsSessionContract(t *testing.T) {
	e := newAuthEnv(t, false)
	rec := post(e.router, "/v1/auth/login", validLoginBody, nil)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, fmt.Sprintf(
		`{"access_token":"a","refresh_token":"r","expires_in":900,"user":{"id":%q,"nombre":"Ana","rol":"cajero"}}`,
		e.svc.session.User.ID), rec.Body.String())
}

func TestLoginHandler_InvalidInput_Returns400WithoutCallingService(t *testing.T) {
	cases := map[string]string{
		"not json":       `{`,
		"empty":          ``,
		"unknown field":  `{"username":"ana","password":"x","rol":"admin"}`,
		"trailing data":  validLoginBody + `{}`,
		"short username": `{"username":"an","password":"x"}`,
		"blank username": `{"username":"     ","password":"x"}`,
		"long username":  fmt.Sprintf(`{"username":%q,"password":"x"}`, strings.Repeat("a", 51)),
		"empty password": `{"username":"ana","password":""}`,
		"long password":  fmt.Sprintf(`{"username":"ana","password":%q}`, strings.Repeat("a", 129)),
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			e := newAuthEnv(t, false)
			rec := post(e.router, "/v1/auth/login", body, nil)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), "INVALID_REQUEST")
			assert.Zero(t, e.svc.loginCalls)
		})
	}
}

func TestLoginHandler_BadCredentials_Returns401(t *testing.T) {
	e := newAuthEnv(t, false)
	e.svc.err = fmt.Errorf("wrapped: %w", domain.ErrInvalidCredentials)
	rec := post(e.router, "/v1/auth/login", validLoginBody, nil)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "INVALID_CREDENTIALS")
}

func TestLoginHandler_InternalError_Returns500GenericAndNoSecretsLogged(t *testing.T) {
	e := newAuthEnv(t, false)
	e.svc.err = errors.New("pg: connection refused 10.0.0.5")
	rec := post(e.router, "/v1/auth/login", validLoginBody, nil)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "10.0.0.5")
	assert.Contains(t, e.logs.String(), "10.0.0.5")
	assert.NotContains(t, e.logs.String(), "clave-secreta-123")
}

func TestLoginHandler_SixthAttempt_Returns429PerIPAndUser(t *testing.T) {
	e := newAuthEnv(t, false)
	e.svc.err = domain.ErrInvalidCredentials
	for range loginRateLimit {
		require.Equal(t, http.StatusUnauthorized, post(e.router, "/v1/auth/login", validLoginBody, nil).Code)
	}
	rec := post(e.router, "/v1/auth/login", `{"username":"ANA","password":"otra"}`, nil)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Contains(t, rec.Body.String(), "RATE_LIMITED")
	assert.Equal(t, loginRateLimit, e.svc.loginCalls)

	other := post(e.router, "/v1/auth/login", `{"username":"beto","password":"x"}`, nil)
	assert.Equal(t, http.StatusUnauthorized, other.Code, "another user has its own bucket")
}

func TestLoginHandler_RateLimit_ForwardedHeaderHonoredOnlyWhenTrusted(t *testing.T) {
	attempts := func(e authEnv) int {
		var last int
		for i := range loginRateLimit + 1 {
			headers := map[string]string{"X-Forwarded-For": fmt.Sprintf("203.0.113.%d", i)}
			last = post(e.router, "/v1/auth/login", validLoginBody, headers).Code
		}
		return last
	}
	assert.Equal(t, http.StatusTooManyRequests, attempts(newAuthEnv(t, false)), "untrusted header cannot evade the limit")
	assert.Equal(t, http.StatusOK, attempts(newAuthEnv(t, true)), "trusted proxy separates clients")
}

func TestClientIP_UsesLastForwardedEntry(t *testing.T) {
	cases := []struct {
		name, xff, want string
		trust           bool
	}{
		{"untrusted", "1.1.1.1", "192.0.2.1", false},
		{"trusted", "1.1.1.1", "1.1.1.1", true},
		{"client-supplied prefix ignored", "9.9.9.9, 1.1.1.1", "1.1.1.1", true},
		{"invalid falls back", "nonsense", "192.0.2.1", true},
		{"absent falls back", "", "192.0.2.1", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
			r.RemoteAddr = "192.0.2.1:1234"
			if c.xff != "" {
				r.Header.Set("X-Forwarded-For", c.xff)
			}
			assert.Equal(t, c.want, clientIP(r, c.trust))
		})
	}
}

func TestRefreshHandler_Responses(t *testing.T) {
	e := newAuthEnv(t, false)
	assert.Equal(t, http.StatusOK, post(e.router, "/v1/auth/refresh", `{"refresh_token":"x"}`, nil).Code)
	assert.Equal(t, http.StatusBadRequest, post(e.router, "/v1/auth/refresh", `{"refresh_token":""}`, nil).Code)
	long := fmt.Sprintf(`{"refresh_token":%q}`, strings.Repeat("a", maxRefreshBytes+1))
	assert.Equal(t, http.StatusBadRequest, post(e.router, "/v1/auth/refresh", long, nil).Code)

	e.svc.err = domain.ErrInvalidRefreshToken
	rec := post(e.router, "/v1/auth/refresh", `{"refresh_token":"x"}`, nil)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "INVALID_REFRESH_TOKEN")
}

func TestLogoutHandler_RequiresAuthAndUsesTokenIdentity(t *testing.T) {
	e := newAuthEnv(t, false)
	body := `{"refresh_token":"x"}`

	rec := post(e.router, "/v1/auth/logout", body, nil)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "Bearer", rec.Header().Get("WWW-Authenticate"))

	id := uuid.New()
	token, err := e.tokens.IssueAccess(domain.Principal{UserID: id, Rol: domain.RoleCajero})
	require.NoError(t, err)
	headers := map[string]string{"Authorization": "Bearer " + token}
	assert.Equal(t, http.StatusNoContent, post(e.router, "/v1/auth/logout", body, headers).Code)
	assert.Equal(t, id, e.svc.logoutUser)
	assert.Equal(t, http.StatusBadRequest, post(e.router, "/v1/auth/logout", `{}`, headers).Code)
}

func TestClientes_Authorization(t *testing.T) {
	e := newAuthEnv(t, false)
	request := func(authz string) *httptest.ResponseRecorder {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/clientes/no-es-uuid/consumos", nil)
		if authz != "" {
			req.Header.Set("Authorization", authz)
		}
		rec := httptest.NewRecorder()
		e.router.ServeHTTP(rec, req)
		return rec
	}

	assert.Equal(t, http.StatusUnauthorized, request("").Code)
	assert.Equal(t, http.StatusUnauthorized, request("Basic abc").Code)
	assert.Equal(t, http.StatusUnauthorized, request("Bearer basura").Code)

	cajero := request("Bearer " + e.bearer(t, domain.RoleCajero))
	assert.Equal(t, http.StatusForbidden, cajero.Code)
	assert.Contains(t, cajero.Body.String(), "FORBIDDEN")

	assert.Equal(t, http.StatusNotFound, request("Bearer "+e.bearer(t, domain.RoleAdmin)).Code, "admin passes authorization")

	expired := e.bearer(t, domain.RoleAdmin)
	e.clock.t = e.clock.t.Add(auth.AccessTokenTTL)
	assert.Equal(t, http.StatusUnauthorized, request("Bearer "+expired).Code)
}

func TestRequireRole_WithoutPrincipal_Returns401(t *testing.T) {
	h := RequireRole(domain.RoleAdmin)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	assert.Equal(t, http.StatusUnauthorized, do(h, http.MethodGet, "/", nil).Code)
}
