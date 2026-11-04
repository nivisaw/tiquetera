// Package http wires the HTTP transport: router, middleware and handlers.
package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"tiquetera/internal/domain"
)

const handlerTimeout = 10 * time.Second

// Deps groups what the router needs from the outside world.
type Deps struct {
	Log    *slog.Logger
	DB     Pinger
	Auth   AuthService
	Tokens TokenVerifier
	// Clientes serves client management.
	Clientes ClienteService
	// Consumos serves lunch debits.
	Consumos ConsumoService
	// Usuarios serves account management.
	Usuarios UsuarioService
	// TrustProxyHeaders enables X-Forwarded-For; set it only behind a proxy.
	TrustProxyHeaders bool
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

	r.Route("/v1", func(v1 chi.Router) {
		auth := newAuthHandler(d)
		v1.Post("/auth/login", auth.login)
		v1.Post("/auth/refresh", auth.refresh)

		v1.Group(func(private chi.Router) {
			private.Use(authenticate(d.Log, d.Tokens, d.Auth))
			private.Post("/auth/logout", auth.logout)

			private.Group(func(cashier chi.Router) {
				cashier.Use(RequireRole(domain.RoleCajero, domain.RoleAdmin), limitPerUser(consumoRateLimit, consumoRateWindow))
				consumos := &consumosHandler{log: d.Log, svc: d.Consumos}
				cashier.Post("/consumos", consumos.create)
			})

			private.Group(func(admin chi.Router) {
				admin.Use(RequireRole(domain.RoleAdmin))
				clientes := &clientesHandler{log: d.Log, svc: d.Clientes}
				admin.Post("/clientes", clientes.create)
				admin.Get("/clientes", clientes.list)
				admin.Get("/clientes/{id}", clientes.get)
				admin.Patch("/clientes/{id}", clientes.update)
				admin.Post("/clientes/{id}/tiqueteras", clientes.recharge)
				admin.Post("/clientes/{id}/qr/regenerar", clientes.regenerateQR)
				admin.Get("/clientes/{id}/consumos", clientes.listConsumos)

				usuarios := &usuariosHandler{log: d.Log, svc: d.Usuarios}
				admin.Post("/usuarios", usuarios.create)
				admin.Get("/usuarios", usuarios.list)
				admin.Get("/usuarios/{id}", usuarios.get)
				admin.Patch("/usuarios/{id}", usuarios.update)
				admin.Put("/usuarios/{id}/password", usuarios.resetPassword)
			})
		})
	})

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
