package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseQRPayload(t *testing.T) {
	payload, hash := NewQRToken()
	token, ok := ParseQRPayload(payload)
	assert.True(t, ok)
	assert.Equal(t, hash, HashQRToken(token))

	valid := strings.TrimPrefix(payload, "TQ1.")
	for name, in := range map[string]string{
		"empty":          "",
		"no prefix":      valid,
		"wrong prefix":   "TQ2." + valid,
		"short":          "TQ1.abc",
		"long":           "TQ1." + valid + "AAAA",
		"not base64":     "TQ1." + strings.Repeat("*", 43),
		"padded base64":  "TQ1." + valid + "=",
		"non-canonical":  "TQ1." + strings.Repeat("A", 42) + "B",
		"prefix only":    "TQ1.",
		"lowercase kind": "tq1." + valid,
	} {
		t.Run(name, func(t *testing.T) {
			_, ok := ParseQRPayload(in)
			assert.False(t, ok)
		})
	}
}

func TestDiagnoseRejection(t *testing.T) {
	today := time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
	active := Cliente{Activo: true}
	book := func(saldo int32, expires time.Time) *Tiquetera {
		return &Tiquetera{Saldo: saldo, FechaVencimiento: expires}
	}
	tests := []struct {
		name    string
		cliente Cliente
		book    *Tiquetera
		code    RejectionCode
		message string
	}{
		{"inactive wins", Cliente{}, book(0, today.AddDate(0, 0, -1)), RejectClienteInactive, "El cliente está inactivo"},
		{"no book", active, nil, RejectNoTiquetera, "El cliente no tiene una tiquetera activa"},
		{"expired wins over balance", active, book(0, today.AddDate(0, 0, -1)), RejectTiqueteraExpired, "Tiquetera vencida el 03/11/2026"},
		{"no balance", active, book(0, today), RejectNoBalance, "La tiquetera no tiene almuerzos disponibles"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, rejected := DiagnoseRejection(tt.cliente, tt.book, today)
			assert.True(t, rejected)
			assert.Equal(t, tt.code, got.Code)
			assert.Equal(t, tt.message, got.Message)
			assert.Contains(t, got.Error(), string(tt.code))
		})
	}

	_, rejected := DiagnoseRejection(active, book(1, today), today)
	assert.False(t, rejected, "an eligible book expiring today has nothing to reject")
}

func TestNewConsumoRejection_QRCodesHaveMessages(t *testing.T) {
	for _, code := range []RejectionCode{RejectQRInvalid, RejectClienteNotFound} {
		assert.NotEmpty(t, NewConsumoRejection(code).Message)
	}
}
