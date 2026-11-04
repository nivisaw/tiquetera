package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tiquetera/internal/domain"
)

type fixedClock struct{ t time.Time }

func (c *fixedClock) Now() time.Time { return c.t }

func newManager(t *testing.T) (*TokenManager, *fixedClock, ed25519.PrivateKey) {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	clock := &fixedClock{t: time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)}
	return NewTokenManager(key, clock), clock, key
}

func sign(t *testing.T, key ed25519.PrivateKey, claims accessClaims) string {
	t.Helper()
	s, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(key)
	require.NoError(t, err)
	return s
}

func TestVerifyAccess_ValidToken_ReturnsPrincipal(t *testing.T) {
	m, _, _ := newManager(t)
	want := domain.Principal{UserID: uuid.New(), Rol: domain.RoleCajero}
	token, err := m.IssueAccess(want)
	require.NoError(t, err)

	got, err := m.VerifyAccess(token)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestVerifyAccess_Expiry_RejectedAfter15Minutes(t *testing.T) {
	m, clock, _ := newManager(t)
	token, err := m.IssueAccess(domain.Principal{UserID: uuid.New(), Rol: domain.RoleAdmin})
	require.NoError(t, err)

	clock.t = clock.t.Add(AccessTokenTTL - time.Second)
	_, err = m.VerifyAccess(token)
	require.NoError(t, err)

	clock.t = clock.t.Add(time.Second)
	_, err = m.VerifyAccess(token)
	assert.ErrorIs(t, err, domain.ErrInvalidAccessToken)
}

func TestVerifyAccess_ForgedTokens_Rejected(t *testing.T) {
	m, clock, key := newManager(t)
	_, otherKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	valid := func() accessClaims {
		return accessClaims{Rol: domain.RoleAdmin, RegisteredClaims: jwt.RegisteredClaims{
			Issuer: issuer, Subject: uuid.NewString(), Audience: jwt.ClaimStrings{audience},
			ExpiresAt: jwt.NewNumericDate(clock.t.Add(time.Minute)),
		}}
	}
	mutate := func(fn func(*accessClaims)) accessClaims {
		c := valid()
		fn(&c)
		return c
	}
	none, err := jwt.NewWithClaims(jwt.SigningMethodNone, valid()).SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)
	public, ok := key.Public().(ed25519.PublicKey)
	require.True(t, ok)
	hmac, err := jwt.NewWithClaims(jwt.SigningMethodHS256, valid()).SignedString([]byte(public))
	require.NoError(t, err)

	cases := map[string]string{
		"garbage":      "abc",
		"alg none":     none,
		"alg confused": hmac,
		"other key":    sign(t, otherKey, valid()),
		"issuer":       sign(t, key, mutate(func(c *accessClaims) { c.Issuer = "x" })),
		"audience":     sign(t, key, mutate(func(c *accessClaims) { c.Audience = jwt.ClaimStrings{"x"} })),
		"no expiry":    sign(t, key, mutate(func(c *accessClaims) { c.ExpiresAt = nil })),
		"subject":      sign(t, key, mutate(func(c *accessClaims) { c.Subject = "1" })),
		"role":         sign(t, key, mutate(func(c *accessClaims) { c.Rol = "root" })),
	}
	for name, token := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := m.VerifyAccess(token)
			assert.ErrorIs(t, err, domain.ErrInvalidAccessToken)
		})
	}
}

func TestNewRefreshToken_ReturnsRandomTokenAndItsHash(t *testing.T) {
	raw1, hash1, err := NewRefreshToken()
	require.NoError(t, err)
	raw2, _, err := NewRefreshToken()
	require.NoError(t, err)

	assert.NotEqual(t, raw1, raw2)
	assert.Len(t, raw1, 43)
	assert.Len(t, hash1, 32)
	assert.Equal(t, HashRefreshToken(raw1), hash1)
}
