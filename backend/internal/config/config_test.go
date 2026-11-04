package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJWTKey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

func env(m map[string]string) Getenv {
	return func(k string) string { return m[k] }
}

func TestLoad_JWTKey_ValidatesSeedWithoutEchoingIt(t *testing.T) {
	for name, value := range map[string]string{"missing": "", "not base64": "@@@", "short": "AAAA"} {
		t.Run(name, func(t *testing.T) {
			_, err := Load(env(map[string]string{"DATABASE_URL": "postgres://x", "JWT_PRIVATE_KEY": value}))
			require.ErrorContains(t, err, "JWT_PRIVATE_KEY")
			if value != "" {
				assert.NotContains(t, err.Error(), value)
			}
		})
	}
}

func TestLoad_JWTKeyFromFile_Valid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "jwt")
	require.NoError(t, os.WriteFile(path, []byte(testJWTKey+"\n"), 0o600))
	cfg, err := Load(env(map[string]string{"DATABASE_URL": "postgres://x", "JWT_PRIVATE_KEY_FILE": path}))
	require.NoError(t, err)
	assert.Len(t, cfg.JWTPrivateKey, 64)
}

func TestLoad_TrustProxyHeaders_Parsed(t *testing.T) {
	cfg, err := Load(env(map[string]string{"DATABASE_URL": "postgres://x", "JWT_PRIVATE_KEY": testJWTKey, "TRUST_PROXY_HEADERS": "true"}))
	require.NoError(t, err)
	assert.True(t, cfg.TrustProxyHeaders)

	_, err = Load(env(map[string]string{"DATABASE_URL": "postgres://x", "JWT_PRIVATE_KEY": testJWTKey, "TRUST_PROXY_HEADERS": "quizas"}))
	assert.ErrorContains(t, err, "TRUST_PROXY_HEADERS")
}

func TestLoad_Defaults_Valid(t *testing.T) {
	cfg, err := Load(env(map[string]string{"DATABASE_URL": "postgres://x", "JWT_PRIVATE_KEY": testJWTKey}))
	require.NoError(t, err)
	assert.Equal(t, "dev", cfg.Env)
	assert.Equal(t, ":8080", cfg.HTTPAddr)
	assert.Equal(t, int32(10), cfg.DBMaxConns)
	assert.Equal(t, slog.LevelInfo, cfg.LogLevel)
}

func TestLoad_Invalid_ReturnsAllErrors(t *testing.T) {
	_, err := Load(env(map[string]string{
		"APP_ENV":      "staging",
		"DB_MAX_CONNS": "0",
		"LOG_LEVEL":    "verbose",
	}))
	require.Error(t, err)
	for _, s := range []string{"APP_ENV", "DATABASE_URL", "DB_MAX_CONNS", "LOG_LEVEL"} {
		assert.Contains(t, err.Error(), s)
	}
}

func TestLoad_SecretFromFile_TrimsValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db")
	require.NoError(t, os.WriteFile(path, []byte("postgres://file\n"), 0o600))

	cfg, err := Load(env(map[string]string{"DATABASE_URL_FILE": path, "JWT_PRIVATE_KEY": testJWTKey}))
	require.NoError(t, err)
	assert.Equal(t, "postgres://file", cfg.DatabaseURL)
}

func TestLoad_SecretFileMissing_Fails(t *testing.T) {
	_, err := Load(env(map[string]string{"DATABASE_URL_FILE": "/nonexistent"}))
	assert.ErrorContains(t, err, "DATABASE_URL_FILE")
}
