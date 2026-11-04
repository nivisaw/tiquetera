package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeTelefono(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"national", "3001234567", "+573001234567"},
		{"national with separators", "(300) 123-4567", "+573001234567"},
		{"country code without plus", "573001234567", "+573001234567"},
		{"colombian e164", "+57 300 123 4567", "+573001234567"},
		{"foreign e164", "+14155552671", "+14155552671"},
		{"too short", "300123", ""},
		{"short colombian e164", "+57300123", ""},
		{"letters", "30012345ab", ""},
		{"zero country code", "+0123456789", ""},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeTelefono(tt.in)
			if tt.want == "" {
				assert.ErrorAs(t, err, &ValidationError{})
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClienteDataNormalize(t *testing.T) {
	text := func(s string) *string { return &s }
	valid := ClienteData{Nombre: " Ana ", Telefono: "3001234567", Documento: text(" 123 "), Email: text("ana@example.com")}

	got, err := valid.Normalize()
	require.NoError(t, err)
	assert.Equal(t, "Ana", got.Nombre)
	assert.Equal(t, "123", *got.Documento)

	tests := []struct {
		name   string
		mutate func(*ClienteData)
	}{
		{"short name", func(d *ClienteData) { d.Nombre = " a " }},
		{"long name", func(d *ClienteData) { d.Nombre = strings.Repeat("ñ", 121) }},
		{"control char in name", func(d *ClienteData) { d.Nombre = "An\x00a" }},
		{"bad phone", func(d *ClienteData) { d.Telefono = "1" }},
		{"empty document", func(d *ClienteData) { d.Documento = text(" ") }},
		{"long document", func(d *ClienteData) { d.Documento = text(strings.Repeat("1", 31)) }},
		{"bad email", func(d *ClienteData) { d.Email = text("no-es-correo") }},
		{"display name email", func(d *ClienteData) { d.Email = text("Ana <ana@example.com>") }},
		{"long email", func(d *ClienteData) { d.Email = text(strings.Repeat("a", 250) + "@b.co") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := valid
			tt.mutate(&d)
			_, err := d.Normalize()
			assert.ErrorAs(t, err, &ValidationError{})
		})
	}
}

func TestClientePatchApply_OnlyChangesSetFields(t *testing.T) {
	doc, nombre := "9", "Nuevo"
	base := ClienteData{Nombre: "Viejo", Telefono: "+573001234567", Documento: &doc, Email: &doc}

	assert.Equal(t, base, ClientePatch{}.Apply(base))

	got := ClientePatch{Nombre: &nombre, SetDocumento: true, SetEmail: true}.Apply(base)
	assert.Equal(t, "Nuevo", got.Nombre)
	assert.Nil(t, got.Documento)
	assert.Nil(t, got.Email)

	phone := "+573009999999"
	assert.Equal(t, phone, ClientePatch{Telefono: &phone}.Apply(base).Telefono)
}

func TestTiqueteraInputValidate(t *testing.T) {
	// 22:00 on Nov 4 in Bogotá, already Nov 5 in UTC.
	now := time.Date(2026, 11, 5, 3, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		almuerzos int
		fecha     string
		ok        bool
	}{
		{"today in Bogotá", 1, "2026-11-04", true},
		{"max lunches", 500, "2027-01-01", true},
		{"yesterday in Bogotá", 1, "2026-11-03", false},
		{"zero lunches", 0, "2026-11-04", false},
		{"too many lunches", 501, "2026-11-04", false},
		{"bad format", 1, "04/11/2026", false},
		{"impossible date", 1, "2026-02-30", false},
		{"missing date", 1, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TiqueteraInput{Almuerzos: tt.almuerzos, FechaVencimiento: tt.fecha}.Validate(now)
			if !tt.ok {
				assert.ErrorAs(t, err, &ValidationError{})
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.fecha, got.Format(time.DateOnly))
		})
	}
}

func TestNewQRToken_HasFormatAndUniqueHash(t *testing.T) {
	payload, hash := NewQRToken()
	token, found := strings.CutPrefix(payload, "TQ1.")
	require.True(t, found)
	assert.Len(t, token, 43)
	assert.Equal(t, HashQRToken(token), hash)
	assert.Len(t, hash, 32)

	other, _ := NewQRToken()
	assert.NotEqual(t, payload, other)
}

func TestValidationError_ReturnsMessage(t *testing.T) {
	assert.Equal(t, "mal", ValidationError{"mal"}.Error())
}
