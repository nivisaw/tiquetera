package domain

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	minPasswordRunes = 10
	maxPasswordRunes = 128
)

// Audit actions for user management.
const (
	AuditUsuarioCreated  = "usuario_creado"
	AuditUsuarioUpdated  = "usuario_actualizado"
	AuditPasswordChanged = "password_restablecido"
)

// User management errors.
var (
	ErrUsernameTaken = errors.New("username already registered")
	ErrLastAdmin     = errors.New("last active admin")
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]{3,50}$`)

// NewUser is a requested account, before validation.
type NewUser struct {
	Username string
	Password string
	Nombre   string
	Rol      Role
}

// Normalize validates the account and returns it in canonical form.
func (u NewUser) Normalize() (NewUser, error) {
	u.Username = strings.TrimSpace(u.Username)
	if !usernamePattern.MatchString(u.Username) {
		return NewUser{}, ValidationError{Message: "El usuario debe tener entre 3 y 50 letras, números, puntos, guiones o guiones bajos"}
	}
	if err := ValidatePassword(u.Password); err != nil {
		return NewUser{}, err
	}
	var err error
	if u.Nombre, err = NormalizeNombre(u.Nombre); err != nil {
		return NewUser{}, err
	}
	if !u.Rol.Valid() {
		return NewUser{}, errInvalidRole()
	}
	return u, nil
}

// ValidatePassword enforces the password length policy.
func ValidatePassword(password string) error {
	n := utf8.RuneCountInString(password)
	if !utf8.ValidString(password) || n < minPasswordRunes || n > maxPasswordRunes {
		return ValidationError{Message: "La contraseña debe tener entre 10 y 128 caracteres"}
	}
	return nil
}

// UserPatch is a partial update of an account.
type UserPatch struct {
	Nombre *string
	Rol    *Role
	Activo *bool
}

// Normalize validates the patch and returns it in canonical form.
func (p UserPatch) Normalize() (UserPatch, error) {
	if p.Nombre != nil {
		nombre, err := NormalizeNombre(*p.Nombre)
		if err != nil {
			return UserPatch{}, err
		}
		p.Nombre = &nombre
	}
	if p.Rol != nil && !p.Rol.Valid() {
		return UserPatch{}, errInvalidRole()
	}
	return p, nil
}

// ChangesAccess reports whether the patch may alter what the account can do.
func (p UserPatch) ChangesAccess() bool { return p.Rol != nil || p.Activo != nil }

// Apply returns u with the patch applied.
func (p UserPatch) Apply(u User) User {
	if p.Nombre != nil {
		u.Nombre = *p.Nombre
	}
	if p.Rol != nil {
		u.Rol = *p.Rol
	}
	if p.Activo != nil {
		u.Activo = *p.Activo
	}
	return u
}

// IsActiveAdmin reports whether the account can currently administer the system.
func (u User) IsActiveAdmin() bool { return u.Activo && u.Rol == RoleAdmin }

func errInvalidRole() error {
	return ValidationError{Message: "El rol debe ser admin o cajero"}
}
