package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"tiquetera/internal/domain"
)

// writeError maps domain errors to HTTP. Anything unknown is logged and
// answered with a generic 500.
func writeError(log *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	var invalid domain.ValidationError
	var rejection domain.ConsumoRejection
	switch {
	case errors.As(err, &rejection):
		writeProblem(w, http.StatusUnprocessableEntity, string(rejection.Code), rejection.Message)
	case errors.Is(err, domain.ErrIdempotencyConflict):
		writeProblem(w, http.StatusConflict, "IDEMPOTENCY_KEY_CONFLICT", "La clave de la solicitud ya fue usada")
	case errors.As(err, &invalid):
		writeProblem(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", invalid.Message)
	case errors.Is(err, domain.ErrNotFound):
		writeProblem(w, http.StatusNotFound, "NOT_FOUND", "Recurso no encontrado")
	case errors.Is(err, domain.ErrPhoneTaken):
		writeProblem(w, http.StatusConflict, "TELEFONO_DUPLICADO", "Ya existe un cliente con ese teléfono")
	case errors.Is(err, domain.ErrClienteInactive):
		writeProblem(w, http.StatusUnprocessableEntity, "CLIENTE_INACTIVO", "El cliente está inactivo")
	case errors.Is(err, domain.ErrUsernameTaken):
		writeProblem(w, http.StatusConflict, "USERNAME_DUPLICADO", "Ya existe un usuario con ese nombre de usuario")
	case errors.Is(err, domain.ErrLastAdmin):
		writeProblem(w, http.StatusUnprocessableEntity, "ULTIMO_ADMIN", "Debe quedar al menos un administrador activo")
	case errors.Is(err, domain.ErrInvalidCredentials):
		writeProblem(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Usuario o contraseña incorrectos, o cuenta bloqueada temporalmente")
	case errors.Is(err, domain.ErrInvalidRefreshToken):
		writeProblem(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "La sesión expiró, inicia sesión de nuevo")
	default:
		log.ErrorContext(r.Context(), "request failed",
			"request_id", middleware.GetReqID(r.Context()), "error", err.Error())
		writeProblem(w, http.StatusInternalServerError, "INTERNAL", "Error interno")
	}
}
