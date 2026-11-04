package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Business limits for clients and ticket books.
const (
	MinAlmuerzos = 1
	MaxAlmuerzos = 500

	minNombreRunes    = 2
	maxNombreRunes    = 120
	maxDocumentoRunes = 30
	minEmailBytes     = 3
	maxEmailBytes     = 254

	bogotaOffsetSeconds = -5 * 60 * 60
	colombiaPrefix      = "+57"
	colombiaPhoneBytes  = 13
	nationalDigits      = 10
	qrTokenBytes        = 32
	qrPayloadPrefix     = "TQ1."
)

// Audit actions for client management.
const (
	AuditClienteCreated   = "cliente_creado"
	AuditClienteUpdated   = "cliente_actualizado"
	AuditTiqueteraCreated = "tiquetera_creada"
	AuditQRRegenerated    = "qr_regenerado"
	AuditEntityCliente    = "cliente"
	AuditEntityTiquetera  = "tiquetera"
)

// Client errors.
var (
	ErrPhoneTaken      = errors.New("phone already registered")
	ErrClienteInactive = errors.New("client is inactive")
)

var e164Pattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

// ValidationError is a rejected input; Message is safe to show to the user.
type ValidationError struct{ Message string }

func (e ValidationError) Error() string { return e.Message }

// ClienteData holds the editable fields of a client.
type ClienteData struct {
	Nombre    string
	Telefono  string
	Documento *string
	Email     *string
}

// Cliente is a registered client. It never carries the QR token or its hash.
type Cliente struct {
	ClienteData
	ID        uuid.UUID
	Activo    bool
	QRVersion int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Tiquetera is a prepaid lunch book.
type Tiquetera struct {
	ID                 uuid.UUID
	ClienteID          uuid.UUID
	AlmuerzosComprados int32
	Saldo              int32
	FechaVencimiento   time.Time
	Activa             bool
	CreatedAt          time.Time
}

// TiqueteraInput is a requested ticket book, before validation.
type TiqueteraInput struct {
	Almuerzos        int
	FechaVencimiento string
}

// ClientePatch is a partial update. The Set flags tell a null value, which
// clears the field, from an absent one.
type ClientePatch struct {
	Nombre       *string
	Telefono     *string
	SetDocumento bool
	Documento    *string
	SetEmail     bool
	Email        *string
	Activo       *bool
}

// Normalize validates the data and returns it in canonical form.
func (d ClienteData) Normalize() (ClienteData, error) {
	var err error
	if d.Nombre, err = NormalizeNombre(d.Nombre); err != nil {
		return ClienteData{}, err
	}
	if d.Telefono, err = NormalizeTelefono(d.Telefono); err != nil {
		return ClienteData{}, err
	}
	if d.Documento, err = normalizeDocumento(d.Documento); err != nil {
		return ClienteData{}, err
	}
	if d.Email, err = normalizeEmail(d.Email); err != nil {
		return ClienteData{}, err
	}
	return d, nil
}

// Apply returns the client's data with the patch applied.
func (p ClientePatch) Apply(d ClienteData) ClienteData {
	if p.Nombre != nil {
		d.Nombre = *p.Nombre
	}
	if p.Telefono != nil {
		d.Telefono = *p.Telefono
	}
	if p.SetDocumento {
		d.Documento = p.Documento
	}
	if p.SetEmail {
		d.Email = p.Email
	}
	return d
}

// NormalizeNombre trims and validates a client name.
func NormalizeNombre(raw string) (string, error) {
	nombre := strings.TrimSpace(raw)
	n := utf8.RuneCountInString(nombre)
	if n < minNombreRunes || n > maxNombreRunes || hasControl(nombre) {
		return "", ValidationError{"El nombre debe tener entre 2 y 120 caracteres"}
	}
	return nombre, nil
}

// NormalizeTelefono returns the phone in E.164. Numbers without a country code
// are taken as Colombian.
func NormalizeTelefono(raw string) (string, error) {
	invalid := ValidationError{"El teléfono no es válido"}
	phone := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "", ".", "").Replace(raw)
	if !strings.HasPrefix(phone, "+") {
		switch {
		case len(phone) == nationalDigits:
			phone = colombiaPrefix + phone
		case len(phone) == colombiaPhoneBytes-1 && strings.HasPrefix(phone, colombiaPrefix[1:]):
			phone = "+" + phone
		default:
			return "", invalid
		}
	}
	if !e164Pattern.MatchString(phone) {
		return "", invalid
	}
	if strings.HasPrefix(phone, colombiaPrefix) && len(phone) != colombiaPhoneBytes {
		return "", invalid
	}
	return phone, nil
}

func normalizeDocumento(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil //nolint:nilnil // an absent optional field is not an error
	}
	documento := strings.TrimSpace(*raw)
	n := utf8.RuneCountInString(documento)
	if n < 1 || n > maxDocumentoRunes || hasControl(documento) {
		return nil, ValidationError{"El documento debe tener entre 1 y 30 caracteres"}
	}
	return &documento, nil
}

func normalizeEmail(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil //nolint:nilnil // an absent optional field is not an error
	}
	email := strings.TrimSpace(*raw)
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email || len(email) < minEmailBytes || len(email) > maxEmailBytes {
		return nil, ValidationError{"El correo no es válido"}
	}
	return &email, nil
}

func hasControl(s string) bool {
	return strings.ContainsFunc(s, unicode.IsControl)
}

// Validate checks the amount and expiry date of a ticket book and returns the
// parsed date. The date may not be earlier than today in Bogotá.
func (in TiqueteraInput) Validate(now time.Time) (time.Time, error) {
	if in.Almuerzos < MinAlmuerzos || in.Almuerzos > MaxAlmuerzos {
		return time.Time{}, ValidationError{"Los almuerzos deben estar entre 1 y 500"}
	}
	fecha, err := time.Parse(time.DateOnly, in.FechaVencimiento)
	if err != nil {
		return time.Time{}, ValidationError{"La fecha de vencimiento no es válida"}
	}
	if fecha.Before(TodayInBogota(now)) {
		return time.Time{}, ValidationError{"La fecha de vencimiento no puede ser anterior a hoy"}
	}
	return fecha, nil
}

// TodayInBogota returns the current Bogotá calendar date at midnight UTC, the
// form dates have when read from the database. Colombia has no daylight saving,
// so a fixed offset is exact and does not depend on the image's tzdata.
func TodayInBogota(now time.Time) time.Time {
	y, m, d := now.In(time.FixedZone("America/Bogota", bogotaOffsetSeconds)).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// NewQRToken returns a new QR payload and the hash to store. The payload must
// not be persisted or logged.
func NewQRToken() (payload string, hash []byte) {
	raw := make([]byte, qrTokenBytes)
	_, _ = rand.Read(raw) // crypto/rand.Read never returns an error
	token := base64.RawURLEncoding.EncodeToString(raw)
	return qrPayloadPrefix + token, HashQRToken(token)
}

// HashQRToken returns the SHA-256 stored for a token.
func HashQRToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
