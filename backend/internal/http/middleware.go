package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

const maxBodyBytes = 64 << 10

// securityHeaders sets defensive response headers for a JSON API.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.Set("Cache-Control", "no-store")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

// limitBody caps request body size to mitigate resource exhaustion.
func limitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		next.ServeHTTP(w, r)
	})
}

// recoverer converts panics into a generic 500 and logs them without leaking details to the client.
func recoverer(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer handlePanic(log, w, r)
			next.ServeHTTP(w, r)
		})
	}
}

func handlePanic(log *slog.Logger, w http.ResponseWriter, r *http.Request) {
	rec := recover()
	if rec == nil {
		return
	}
	if rec == http.ErrAbortHandler { //nolint:errorlint // sentinel compared by identity per net/http docs
		panic(rec)
	}
	log.ErrorContext(r.Context(), "panic recovered",
		"request_id", middleware.GetReqID(r.Context()), "panic", rec)
	writeProblem(w, http.StatusInternalServerError, "INTERNAL", "Error interno")
}

// requestLogger logs one line per request. Never logs bodies, query strings or headers.
func requestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			log.InfoContext(r.Context(), "http request",
				"request_id", middleware.GetReqID(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
