package platform

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/argon2"
)

func TestHashPassword_ValidPassword_ReturnsVerifiablePHCString(t *testing.T) {
	hash, err := HashPassword("una-clave-segura")
	require.NoError(t, err)

	parts := strings.Split(hash, "$")
	require.Len(t, parts, 6)
	assert.Equal(t, "argon2id", parts[1])
	assert.Equal(t, "v=19", parts[2])
	assert.Equal(t, "m=65536,t=3,p=2", parts[3])

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	require.NoError(t, err)
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	require.NoError(t, err)
	assert.Equal(t, argon2.IDKey([]byte("una-clave-segura"), salt, 3, 64*1024, 2, 32), key)
}

func TestHashPassword_SamePassword_UsesDifferentSalts(t *testing.T) {
	first, err := HashPassword("una-clave-segura")
	require.NoError(t, err)
	second, err := HashPassword("una-clave-segura")
	require.NoError(t, err)
	assert.NotEqual(t, first, second)
}
