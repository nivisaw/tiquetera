package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func env(m map[string]string) Getenv {
	return func(k string) string { return m[k] }
}

func TestLoad_Defaults_Valid(t *testing.T) {
	cfg, err := Load(env(map[string]string{"DATABASE_URL": "postgres://x"}))
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

	cfg, err := Load(env(map[string]string{"DATABASE_URL_FILE": path}))
	require.NoError(t, err)
	assert.Equal(t, "postgres://file", cfg.DatabaseURL)
}

func TestLoad_SecretFileMissing_Fails(t *testing.T) {
	_, err := Load(env(map[string]string{"DATABASE_URL_FILE": "/nonexistent"}))
	assert.ErrorContains(t, err, "DATABASE_URL_FILE")
}
