// Package config loads and validates runtime configuration from the environment.
package config

import (
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

	dbURL, err := secret(getenv, "DATABASE_URL")
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

	return cfg, errors.Join(errs...)
}

func secret(getenv Getenv, name string) (string, error) {
	if path := getenv(name + "_FILE"); path != "" {
		b, err := os.ReadFile(path) //nolint:gosec // path comes from operator-controlled env
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
