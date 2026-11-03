// Package http wires the HTTP transport: router, middleware and handlers.
package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const handlerTimeout = 10 * time.Second

// Deps groups what the router needs from the outside world.
type Deps struct {
	Log *slog.Logger
	DB  Pinger
}

// NewRouter builds the API router with the standard middleware chain.
func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		requestLogger(d.Log),
		recoverer(d.Log),
		securityHeaders,
		limitBody,
		middleware.Timeout(handlerTimeout),
	)

	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		writeProblem(w, http.StatusNotFound, "NOT_FOUND", "Recurso no encontrado")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		writeProblem(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido")
	})

	r.Get("/healthz", healthz)
	r.Get("/readyz", readyz(d.DB))

	r.Route("/v1", func(_ chi.Router) {})

	return r
}

// NewServer returns an http.Server with hardened timeouts.
func NewServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
}
