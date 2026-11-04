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

func TestVerifyPassword_MatchingAndWrong_ReportsResult(t *testing.T) {
	hash, err := HashPassword("una-clave-segura")
	require.NoError(t, err)

	ok, err := VerifyPassword("una-clave-segura", hash)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = VerifyPassword("otra-clave", hash)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestVerifyPassword_MalformedHash_ReturnsError(t *testing.T) {
	salt := base64.RawStdEncoding.EncodeToString(make([]byte, 16))
	key := base64.RawStdEncoding.EncodeToString(make([]byte, 32))
	cases := map[string]string{
		"empty":       "",
		"algorithm":   "$bcrypt$v=19$m=65536,t=3,p=2$" + salt + "$" + key,
		"version":     "$argon2id$v=1$m=65536,t=3,p=2$" + salt + "$" + key,
		"params":      "$argon2id$v=19$nope$" + salt + "$" + key,
		"zero time":   "$argon2id$v=19$m=65536,t=0,p=2$" + salt + "$" + key,
		"huge memory": "$argon2id$v=19$m=99999999,t=3,p=2$" + salt + "$" + key,
		"salt":        "$argon2id$v=19$m=65536,t=3,p=2$***$" + key,
		"key":         "$argon2id$v=19$m=65536,t=3,p=2$" + salt + "$***",
		"short key":   "$argon2id$v=19$m=65536,t=3,p=2$" + salt + "$AAAA",
	}
	for name, hash := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := VerifyPassword("x", hash)
			assert.Error(t, err)
		})
	}
}

func TestHashPassword_SamePassword_UsesDifferentSalts(t *testing.T) {
	first, err := HashPassword("una-clave-segura")
	require.NoError(t, err)
	second, err := HashPassword("una-clave-segura")
	require.NoError(t, err)
	assert.NotEqual(t, first, second)
}
