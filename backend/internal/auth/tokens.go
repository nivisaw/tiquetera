// Package auth issues and verifies access tokens and generates refresh tokens.
package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"tiquetera/internal/domain"
	"tiquetera/internal/platform"
)

const (
	// AccessTokenTTL is the lifetime of a JWT access token.
	AccessTokenTTL = 15 * time.Minute
	// RefreshTokenTTL is the lifetime of an opaque refresh token.
	RefreshTokenTTL = 30 * 24 * time.Hour

	issuer         = "tiquetera"
	audience       = "tiquetera-api"
	refreshByteLen = 32
)

type accessClaims struct {
	Rol domain.Role `json:"rol"`
	jwt.RegisteredClaims
}

// TokenManager signs and verifies EdDSA access tokens.
type TokenManager struct {
	key    ed25519.PrivateKey
	clock  platform.Clock
	parser *jwt.Parser
}

// NewTokenManager returns a TokenManager using key and clock.
func NewTokenManager(key ed25519.PrivateKey, clock platform.Clock) *TokenManager {
	return &TokenManager{
		key:   key,
		clock: clock,
		parser: jwt.NewParser(
			jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
			jwt.WithIssuer(issuer),
			jwt.WithAudience(audience),
			jwt.WithExpirationRequired(),
			jwt.WithTimeFunc(clock.Now),
		),
	}
}

// IssueAccess returns a signed access token for principal.
func (m *TokenManager) IssueAccess(p domain.Principal) (string, error) {
	now := m.clock.Now()
	claims := accessClaims{
		Rol: p.Rol,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   p.UserID.String(),
			Audience:  jwt.ClaimStrings{audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(m.key)
	if err != nil {
		return "", fmt.Errorf("signing access token: %w", err)
	}
	return signed, nil
}

// VerifyAccess validates an access token and returns its principal.
func (m *TokenManager) VerifyAccess(token string) (domain.Principal, error) {
	var claims accessClaims
	public := m.key.Public()
	_, err := m.parser.ParseWithClaims(token, &claims, func(*jwt.Token) (any, error) { return public, nil })
	if err != nil {
		return domain.Principal{}, domain.ErrInvalidAccessToken
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil || !claims.Rol.Valid() {
		return domain.Principal{}, domain.ErrInvalidAccessToken
	}
	return domain.Principal{UserID: id, Rol: claims.Rol}, nil
}

// NewRefreshToken returns a random opaque token and the hash to store.
func NewRefreshToken() (raw string, hash []byte, err error) {
	b := make([]byte, refreshByteLen)
	if _, err := rand.Read(b); err != nil {
		return "", nil, fmt.Errorf("generating refresh token: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	return raw, HashRefreshToken(raw), nil
}

// HashRefreshToken returns the SHA-256 of a raw refresh token.
func HashRefreshToken(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}
