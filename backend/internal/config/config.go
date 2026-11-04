// Package config loads and validates runtime configuration from the environment.
package config

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds validated runtime settings.
type Config struct {
	Env             string
	HTTPAddr        string
	DatabaseURL     string
	DBMaxConns      int32
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
	JWTPrivateKey   ed25519.PrivateKey
	// TrustProxyHeaders enables X-Forwarded-For; set it only behind a proxy.
	TrustProxyHeaders bool
}

// Getenv abstracts environment lookup for testability.
type Getenv func(string) string

// Load reads configuration and fails fast on invalid or missing values.
// Secrets may be provided as NAME or NAME_FILE (path to a mounted file).
func Load(getenv Getenv) (Config, error) {
	var errs []error

	cfg := Config{
		Env:             valueOr(getenv("APP_ENV"), "dev"),
		HTTPAddr:        valueOr(getenv("HTTP_ADDR"), ":8080"),
		ShutdownTimeout: 15 * time.Second,
	}

	if cfg.Env != "dev" && cfg.Env != "test" && cfg.Env != "prod" {
		errs = append(errs, fmt.Errorf("APP_ENV must be dev, test or prod"))
	}

	dbURL, err := Secret(getenv, "DATABASE_URL")
	if err != nil {
		errs = append(errs, err)
	} else if dbURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	cfg.DatabaseURL = dbURL

	maxConns, err := strconv.ParseInt(valueOr(getenv("DB_MAX_CONNS"), "10"), 10, 32)
	if err != nil || maxConns < 1 || maxConns > 100 {
		errs = append(errs, errors.New("DB_MAX_CONNS must be an integer between 1 and 100"))
	}
	cfg.DBMaxConns = int32(maxConns)

	if err := cfg.LogLevel.UnmarshalText([]byte(valueOr(getenv("LOG_LEVEL"), "info"))); err != nil {
		errs = append(errs, errors.New("LOG_LEVEL must be debug, info, warn or error"))
	}

	cfg.JWTPrivateKey, err = loadJWTKey(getenv)
	if err != nil {
		errs = append(errs, err)
	}

	cfg.TrustProxyHeaders, err = strconv.ParseBool(valueOr(getenv("TRUST_PROXY_HEADERS"), "false"))
	if err != nil {
		errs = append(errs, errors.New("TRUST_PROXY_HEADERS must be true or false"))
	}

	return cfg, errors.Join(errs...)
}

// loadJWTKey reads an Ed25519 seed encoded in base64. The value is never echoed.
func loadJWTKey(getenv Getenv) (ed25519.PrivateKey, error) {
	encoded, err := Secret(getenv, "JWT_PRIVATE_KEY")
	if err != nil {
		return nil, err
	}
	seed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil, errors.New("JWT_PRIVATE_KEY must be a base64-encoded 32-byte seed (openssl rand -base64 32)")
	}
	return ed25519.NewKeyFromSeed(seed), nil
}

// Secret reads NAME_FILE when set, otherwise NAME.
func Secret(getenv Getenv, name string) (string, error) {
	if path := getenv(name + "_FILE"); path != "" {
		b, err := os.ReadFile(path) // #nosec G304 -- path comes from operator-controlled env
		if err != nil {
			return "", fmt.Errorf("reading %s_FILE: %w", name, err)
		}
		return strings.TrimSpace(string(b)), nil
	}
	return getenv(name), nil
}

func valueOr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
