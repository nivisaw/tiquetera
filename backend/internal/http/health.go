package http

import (
	"context"
	"net/http"
	"time"
)

// Pinger reports dependency readiness (e.g. the database pool).
type Pinger interface {
	Ping(ctx context.Context) error
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func readyz(db Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			writeProblem(w, http.StatusServiceUnavailable, "NOT_READY", "Servicio no disponible")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}
