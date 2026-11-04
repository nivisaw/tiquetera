package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validPassword = "clave-segura-123"

func validNewUser() NewUser {
	return NewUser{Username: "  caja.1 ", Password: validPassword, Nombre: " Ana ", Rol: RoleCajero}
}

func TestNewUserNormalize_Valid_TrimsFields(t *testing.T) {
	got, err := validNewUser().Normalize()
	require.NoError(t, err)
	assert.Equal(t, "caja.1", got.Username)
	assert.Equal(t, "Ana", got.Nombre)
	assert.Equal(t, validPassword, got.Password, "password is never altered")
}

func TestNewUserNormalize_Invalid_ReturnsValidationError(t *testing.T) {
	cases := map[string]func(*NewUser){
		"short username":   func(u *NewUser) { u.Username = "ab" },
		"long username":    func(u *NewUser) { u.Username = strings.Repeat("a", 51) },
		"username symbols": func(u *NewUser) { u.Username = "ana pérez" },
		"username nul":     func(u *NewUser) { u.Username = "ana\x00" },
		"short password":   func(u *NewUser) { u.Password = strings.Repeat("a", 9) },
		"long password":    func(u *NewUser) { u.Password = strings.Repeat("a", 129) },
		"invalid utf8":     func(u *NewUser) { u.Password = strings.Repeat("\xff", 10) },
		"short nombre":     func(u *NewUser) { u.Nombre = "A" },
		"unknown role":     func(u *NewUser) { u.Rol = "root" },
		"empty role":       func(u *NewUser) { u.Rol = "" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			u := validNewUser()
			mutate(&u)
			_, err := u.Normalize()
			var invalid ValidationError
			assert.ErrorAs(t, err, &invalid)
		})
	}
}

func TestValidatePassword_Bounds(t *testing.T) {
	require.NoError(t, ValidatePassword(strings.Repeat("ñ", 10)), "counts characters, not bytes")
	require.NoError(t, ValidatePassword(strings.Repeat("a", 128)))
	assert.Error(t, ValidatePassword(strings.Repeat("a", 9)))
}

func TestUserPatch_NormalizeAndApply(t *testing.T) {
	nombre, rol, activo := " Rosa ", RoleAdmin, false
	patch, err := UserPatch{Nombre: &nombre, Rol: &rol, Activo: &activo}.Normalize()
	require.NoError(t, err)
	assert.True(t, patch.ChangesAccess())

	got := patch.Apply(User{Nombre: "Ana", Rol: RoleCajero, Activo: true})
	assert.Equal(t, User{Nombre: "Rosa", Rol: RoleAdmin, Activo: false}, got)
	assert.False(t, got.IsActiveAdmin())
	assert.Equal(t, User{Nombre: "Ana"}, UserPatch{}.Apply(User{Nombre: "Ana"}))
}

func TestUserPatchNormalize_Invalid_ReturnsValidationError(t *testing.T) {
	nombre, rol := "A", Role("root")
	_, err := UserPatch{Nombre: &nombre}.Normalize()
	var invalid ValidationError
	require.ErrorAs(t, err, &invalid)
	_, err = UserPatch{Rol: &rol}.Normalize()
	require.ErrorAs(t, err, &invalid)
	assert.False(t, UserPatch{Nombre: &nombre}.ChangesAccess())
}
