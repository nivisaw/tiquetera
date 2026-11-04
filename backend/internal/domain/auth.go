// Package domain holds entities, domain errors and pure business rules.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	// MaxFailedLogins is the number of consecutive failures that locks an account.
	MaxFailedLogins = 5
	// LockDuration is how long an account stays locked.
	LockDuration = 15 * time.Minute
)

// Role is a server-side authorization role.
type Role string

// Supported roles.
const (
	RoleAdmin  Role = "admin"
	RoleCajero Role = "cajero"
)

// Valid reports whether r is a known role.
func (r Role) Valid() bool { return r == RoleAdmin || r == RoleCajero }

// Domain errors.
var (
	ErrNotFound            = errors.New("not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidAccessToken  = errors.New("invalid access token")
)

// User is an application account.
type User struct {
	ID           uuid.UUID
	Username     string
	Nombre       string
	PasswordHash string
	Rol          Role
	Activo       bool
	FailedLogins int32
	LockedUntil  *time.Time
}

// IsLocked reports whether the account is locked at now.
func (u User) IsLocked(now time.Time) bool {
	return u.LockedUntil != nil && u.LockedUntil.After(now)
}

// WithFailedLogin returns the user after one more failed attempt. An expired
// lock starts a new count; reaching MaxFailedLogins sets a new lock.
func (u User) WithFailedLogin(now time.Time) User {
	if u.LockedUntil != nil && !u.IsLocked(now) {
		u.FailedLogins, u.LockedUntil = 0, nil
	}
	u.FailedLogins++
	if u.FailedLogins >= MaxFailedLogins {
		until := now.Add(LockDuration)
		u.LockedUntil = &until
	}
	return u
}

// Principal is the authenticated identity of a request.
type Principal struct {
	UserID uuid.UUID
	Rol    Role
}

// RefreshToken is a stored refresh token. Only its hash is persisted.
type RefreshToken struct {
	ID        uuid.UUID
	UsuarioID uuid.UUID
	TokenHash []byte
	FamilyID  uuid.UUID
	ExpiresAt time.Time
	RevokedAt *time.Time
}

// Session is the result of a successful login or refresh.
type Session struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	User         User
}

// Audit actions for authentication events.
const (
	AuditLoginOK       = "login_ok"
	AuditLoginFailed   = "login_fallido"
	AuditLoginRejected = "login_rechazado"
	AuditAccountLocked = "cuenta_bloqueada"
	AuditRefreshReused = "refresh_reutilizado"
	AuditLogout        = "logout"
	AuditEntityUsuario = "usuario"
)

// AuditEntry is one audit_log row. It must not carry passwords or tokens.
type AuditEntry struct {
	UsuarioID *uuid.UUID
	Accion    string
	Entidad   string
	EntidadID *uuid.UUID
}
