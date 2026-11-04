package domain

import (
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// RejectionCode is why a lunch could not be debited.
type RejectionCode string

// Rejection codes, as stored in consumos_rechazados and returned by the API.
const (
	RejectQRInvalid        RejectionCode = "QR_INVALIDO"
	RejectClienteNotFound  RejectionCode = "CLIENTE_NO_EXISTE"
	RejectClienteInactive  RejectionCode = "CLIENTE_INACTIVO"
	RejectNoTiquetera      RejectionCode = "SIN_TIQUETERA"
	RejectTiqueteraExpired RejectionCode = "TIQUETERA_VENCIDA"
	RejectNoBalance        RejectionCode = "SIN_SALDO"
)

const displayDate = "02/01/2006"

// Consumption errors.
var (
	// ErrIdempotencyConflict means another user already used the key.
	ErrIdempotencyConflict = errors.New("idempotency key used by another user")
	// ErrDuplicateIdempotencyKey means a concurrent request stored the key first.
	ErrDuplicateIdempotencyKey = errors.New("duplicate idempotency key")
)

// ConsumoRejection is a business rejection; Message is safe to show to the user.
type ConsumoRejection struct {
	Code    RejectionCode
	Message string
}

func (e ConsumoRejection) Error() string { return "consumption rejected: " + string(e.Code) }

// NewConsumoRejection returns the rejection for a code that needs no details.
func NewConsumoRejection(code RejectionCode) ConsumoRejection {
	messages := map[RejectionCode]string{
		RejectQRInvalid:       "El código QR no es válido",
		RejectClienteNotFound: "El código QR no está registrado o fue reemplazado",
		RejectClienteInactive: "El cliente está inactivo",
		RejectNoTiquetera:     "El cliente no tiene una tiquetera activa",
		RejectNoBalance:       "La tiquetera no tiene almuerzos disponibles",
	}
	return ConsumoRejection{Code: code, Message: messages[code]}
}

// ConsumoResult is what a successful debit returns, also on an idempotent replay.
type ConsumoResult struct {
	CajeroID         uuid.UUID
	ClienteNombre    string
	SaldoRestante    int32
	FechaVencimiento time.Time
	FechaHora        time.Time
}

// Consumo is one debit to persist.
type Consumo struct {
	TiqueteraID     uuid.UUID
	ClienteID       uuid.UUID
	CajeroID        uuid.UUID
	SaldoResultante int32
	IdempotencyKey  uuid.UUID
}

// ParseQRPayload returns the token of a well-formed TQ1 payload.
func ParseQRPayload(payload string) (string, bool) {
	token, found := strings.CutPrefix(payload, qrPayloadPrefix)
	if !found {
		return "", false
	}
	raw, err := base64.RawURLEncoding.Strict().DecodeString(token)
	if err != nil || len(raw) != qrTokenBytes {
		return "", false
	}
	return token, true
}

// DiagnoseRejection explains why a debit did not apply. It returns false when
// nothing prevents the debit, which means the state changed since the attempt.
func DiagnoseRejection(c Cliente, t *Tiquetera, today time.Time) (ConsumoRejection, bool) {
	switch {
	case !c.Activo:
		return NewConsumoRejection(RejectClienteInactive), true
	case t == nil:
		return NewConsumoRejection(RejectNoTiquetera), true
	case t.FechaVencimiento.Before(today):
		return ConsumoRejection{
			Code:    RejectTiqueteraExpired,
			Message: "Tiquetera vencida el " + t.FechaVencimiento.Format(displayDate),
		}, true
	case t.Saldo <= 0:
		return NewConsumoRejection(RejectNoBalance), true
	default:
		return ConsumoRejection{}, false
	}
}
