package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

func newTestRouter(db Pinger) http.Handler {
	return NewRouter(Deps{Log: slog.New(slog.NewJSONHandler(io.Discard, nil)), DB: db})
}

func do(h http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), method, path, body))
	return rec
}

func TestHealthz_ReturnsOK(t *testing.T) {
	rec := do(newTestRouter(fakePinger{}), http.MethodGet, "/healthz", nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
}

func TestReadyz_DBUp_ReturnsOK(t *testing.T) {
	rec := do(newTestRouter(fakePinger{}), http.MethodGet, "/readyz", nil)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestReadyz_DBDown_Returns503WithoutDetails(t *testing.T) {
	rec := do(newTestRouter(fakePinger{err: errors.New("dial tcp 10.0.0.5: refused")}), http.MethodGet, "/readyz", nil)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
	assert.NotContains(t, rec.Body.String(), "10.0.0.5")
}

func TestRouter_SetsSecurityHeaders(t *testing.T) {
	rec := do(newTestRouter(fakePinger{}), http.MethodGet, "/healthz", nil)
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
}

func TestRouter_UnknownRoute_Returns404Problem(t *testing.T) {
	rec := do(newTestRouter(fakePinger{}), http.MethodGet, "/nope", nil)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"NOT_FOUND"`)
}

func TestLimitBody_Oversized_FailsRead(t *testing.T) {
	var readErr error
	h := limitBody(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, readErr = io.ReadAll(r.Body)
	}))
	do(h, http.MethodPost, "/", strings.NewReader(strings.Repeat("a", maxBodyBytes+1)))
	var maxErr *http.MaxBytesError
	require.ErrorAs(t, readErr, &maxErr)
}

func TestRecoverer_Panic_Returns500Generic(t *testing.T) {
	h := recoverer(slog.New(slog.NewJSONHandler(io.Discard, nil)))(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("secret internal state") }))
	rec := do(h, http.MethodGet, "/", nil)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "secret")
}
