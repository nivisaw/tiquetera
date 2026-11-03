package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSeedConfig_ValidEnv_ReturnsConfig(t *testing.T) {
	cfg, err := loadSeedConfig(envOf(validEnv()))
	require.NoError(t, err)
	assert.Equal(t, "admin", cfg.username)
	assert.Equal(t, "una-clave-segura", cfg.password)
}

func TestLoadSeedConfig_InvalidEnv_ReturnsError(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
		want  string
	}{
		{"sin url", "DATABASE_URL", "", "DATABASE_URL"},
		{"clave corta", "SEED_ADMIN_PASSWORD", "123456789", "SEED_ADMIN_PASSWORD"},
		{"sin usuario", "SEED_ADMIN_USERNAME", "", "SEED_ADMIN_USERNAME"},
		{"sin nombre", "SEED_ADMIN_NOMBRE", "", "SEED_ADMIN_NOMBRE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := validEnv()
			env[tt.key] = tt.value
			_, err := loadSeedConfig(envOf(env))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func validEnv() map[string]string {
	return map[string]string{
		"DATABASE_URL":        "postgres://app:secret@db/tiquetera",
		"SEED_ADMIN_USERNAME": "admin",
		"SEED_ADMIN_PASSWORD": "una-clave-segura",
		"SEED_ADMIN_NOMBRE":   "Administrador",
	}
}

func envOf(env map[string]string) func(string) string {
	return func(key string) string { return env[key] }
}
