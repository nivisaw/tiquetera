package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/httprate"
	"github.com/google/uuid"

	"tiquetera/internal/domain"
)

const (
	consumoRateLimit  = 60
	consumoRateWindow = time.Minute
	maxQRBytes        = 128
)

// ConsumoService is the consumption use-case port.
type ConsumoService interface {
	Consume(ctx context.Context, cajero, key uuid.UUID, qr string) (domain.ConsumoResult, error)
}

type createConsumoRequest struct {
	QR string `json:"qr"`
}

type consumoCliente struct {
	Nombre string `json:"nombre"`
}

type consumoResponse struct {
	OK               bool           `json:"ok"`
	Cliente          consumoCliente `json:"cliente"`
	SaldoRestante    int32          `json:"saldo_restante"`
	FechaVencimiento string         `json:"fecha_vencimiento"`
	FechaHora        time.Time      `json:"fecha_hora"`
}

type consumosHandler struct {
	log *slog.Logger
	svc ConsumoService
}

func (h *consumosHandler) create(w http.ResponseWriter, r *http.Request) {
	cajero, ok := actorFrom(w, r)
	if !ok {
		return
	}
	key, err := uuid.Parse(r.Header.Get("Idempotency-Key"))
	if err != nil || key == uuid.Nil {
		writeInvalidRequest(w)
		return
	}
	var req createConsumoRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.QR == "" || len(req.QR) > maxQRBytes {
		writeInvalidRequest(w)
		return
	}
	result, err := h.svc.Consume(r.Context(), cajero, key, req.QR)
	if err != nil {
		writeError(h.log, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, consumoResponse{
		OK:               true,
		Cliente:          consumoCliente{Nombre: result.ClienteNombre},
		SaldoRestante:    result.SaldoRestante,
		FechaVencimiento: result.FechaVencimiento.Format(time.DateOnly),
		FechaHora:        result.FechaHora,
	})
}

// limitPerUser rate-limits by the authenticated user, not by address, so
// cashiers sharing a network do not affect each other.
func limitPerUser(limit int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.LimitBy(limit, window,
		func(r *http.Request) (string, error) {
			principal, ok := principalFrom(r.Context())
			if !ok {
				return "", errors.New("missing principal")
			}
			return principal.UserID.String(), nil
		},
		httprate.WithLimitHandler(func(w http.ResponseWriter, _ *http.Request) {
			writeProblem(w, http.StatusTooManyRequests, "RATE_LIMITED", "Demasiadas solicitudes, vuelve a intentarlo en un minuto")
		}))
}
