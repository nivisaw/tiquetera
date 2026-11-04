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
	switch {
	case errors.As(err, &invalid):
		writeProblem(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", invalid.Message)
	case errors.Is(err, domain.ErrNotFound):
		writeProblem(w, http.StatusNotFound, "NOT_FOUND", "Recurso no encontrado")
	case errors.Is(err, domain.ErrPhoneTaken):
		writeProblem(w, http.StatusConflict, "TELEFONO_DUPLICADO", "Ya existe un cliente con ese teléfono")
	case errors.Is(err, domain.ErrClienteInactive):
		writeProblem(w, http.StatusUnprocessableEntity, "CLIENTE_INACTIVO", "El cliente está inactivo")
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
