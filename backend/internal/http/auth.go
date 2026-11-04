package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/httprate"
	"github.com/google/uuid"

	"tiquetera/internal/domain"
)

const (
	loginRateLimit   = 5
	loginRateWindow  = time.Minute
	minUsernameRunes = 3
	maxUsernameRunes = 50
	maxPasswordRunes = 128
	maxRefreshBytes  = 512
)

// AuthService is the authentication use-case port.
type AuthService interface {
	Login(ctx context.Context, username, password string) (domain.Session, error)
	Refresh(ctx context.Context, rawToken string) (domain.Session, error)
	Logout(ctx context.Context, p domain.Principal, rawToken string) error
}

// TokenVerifier validates access tokens.
type TokenVerifier interface {
	VerifyAccess(token string) (domain.Principal, error)
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type sessionUser struct {
	ID     uuid.UUID   `json:"id"`
	Nombre string      `json:"nombre"`
	Rol    domain.Role `json:"rol"`
}

type sessionResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
	User         sessionUser `json:"user"`
}

func toSessionResponse(s domain.Session) sessionResponse {
	return sessionResponse{
		AccessToken:  s.AccessToken,
		RefreshToken: s.RefreshToken,
		ExpiresIn:    s.ExpiresIn,
		User:         sessionUser{ID: s.User.ID, Nombre: s.User.Nombre, Rol: s.User.Rol},
	}
}

type authHandler struct {
	log        *slog.Logger
	svc        AuthService
	limiter    *httprate.RateLimiter
	trustProxy bool
}

func newAuthHandler(d Deps) *authHandler {
	limiter := httprate.NewRateLimiter(loginRateLimit, loginRateWindow,
		httprate.WithLimitHandler(func(w http.ResponseWriter, _ *http.Request) {
			writeProblem(w, http.StatusTooManyRequests, "RATE_LIMITED", "Demasiados intentos, vuelve a intentarlo en un minuto")
		}))
	return &authHandler{log: d.Log, svc: d.Auth, limiter: limiter, trustProxy: d.TrustProxyHeaders}
}

func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if !validLogin(req) {
		writeInvalidRequest(w)
		return
	}
	key := httprate.CanonicalizeIP(clientIP(r, h.trustProxy)) + "|" + strings.ToLower(req.Username)
	if h.limiter.RespondOnLimit(w, r, key) {
		return
	}
	session, err := h.svc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		writeError(h.log, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toSessionResponse(session))
}

func validLogin(req loginRequest) bool {
	username := utf8.RuneCountInString(req.Username)
	password := utf8.RuneCountInString(req.Password)
	return username >= minUsernameRunes && username <= maxUsernameRunes &&
		password >= 1 && password <= maxPasswordRunes
}

func (h *authHandler) refresh(w http.ResponseWriter, r *http.Request) {
	token, ok := decodeRefreshToken(w, r)
	if !ok {
		return
	}
	session, err := h.svc.Refresh(r.Context(), token)
	if err != nil {
		writeError(h.log, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toSessionResponse(session))
}

func (h *authHandler) logout(w http.ResponseWriter, r *http.Request) {
	token, ok := decodeRefreshToken(w, r)
	if !ok {
		return
	}
	principal, ok := principalFrom(r.Context())
	if !ok {
		writeUnauthenticated(w)
		return
	}
	if err := h.svc.Logout(r.Context(), principal, token); err != nil {
		writeError(h.log, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeRefreshToken(w http.ResponseWriter, r *http.Request) (string, bool) {
	var req refreshRequest
	if !decodeJSON(w, r, &req) {
		return "", false
	}
	if req.RefreshToken == "" || len(req.RefreshToken) > maxRefreshBytes {
		writeInvalidRequest(w)
		return "", false
	}
	return req.RefreshToken, true
}

// decodeJSON parses a single strict JSON object, answering 400 when it cannot.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeInvalidRequest(w)
		return false
	}
	if !errors.Is(dec.Decode(&struct{}{}), io.EOF) {
		writeInvalidRequest(w)
		return false
	}
	return true
}

func writeInvalidRequest(w http.ResponseWriter) {
	writeProblem(w, http.StatusBadRequest, "INVALID_REQUEST", "Solicitud inválida")
}

// clientIP returns the caller's address. X-Forwarded-For is honored only behind
// a trusted proxy, and only its last entry: the one that proxy appended.
func clientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		forwarded := r.Header.Values("X-Forwarded-For")
		if len(forwarded) > 0 {
			last := forwarded[len(forwarded)-1]
			last = strings.TrimSpace(last[strings.LastIndexByte(last, ',')+1:])
			if net.ParseIP(last) != nil {
				return last
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type principalKey struct{}

func principalFrom(ctx context.Context) (domain.Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(domain.Principal)
	return p, ok
}

func writeUnauthenticated(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	writeProblem(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Autenticación requerida")
}

// authenticate requires a valid access token and stores its principal.
func authenticate(tokens TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scheme, token, found := strings.Cut(r.Header.Get("Authorization"), " ")
			if !found || !strings.EqualFold(scheme, "Bearer") {
				writeUnauthenticated(w)
				return
			}
			principal, err := tokens.VerifyAccess(strings.TrimSpace(token))
			if err != nil {
				writeUnauthenticated(w)
				return
			}
			ctx := context.WithValue(r.Context(), principalKey{}, principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole allows only the given roles, using the role in the verified token.
func RequireRole(roles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := principalFrom(r.Context())
			if !ok {
				writeUnauthenticated(w)
				return
			}
			if !slices.Contains(roles, principal.Rol) {
				writeProblem(w, http.StatusForbidden, "FORBIDDEN", "No tienes permiso para esta acción")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
