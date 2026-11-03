// Command seed creates the initial admin user if it does not exist yet.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"tiquetera/internal/config"
	"tiquetera/internal/platform"
	"tiquetera/internal/repository/postgres/sqlcgen"
)

const minPasswordLength = 10

type seedConfig struct {
	databaseURL string
	username    string
	password    string
	nombre      string
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := loadSeedConfig(os.Getenv)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	hash, err := platform.HashPassword(cfg.password)
	if err != nil {
		return err
	}
	connCfg, err := pgx.ParseConfig(cfg.databaseURL)
	if err != nil {
		return errors.New("invalid DATABASE_URL") // never echo the URL: it contains credentials
	}
	conn, err := pgx.ConnectConfig(ctx, connCfg)
	if err != nil {
		return fmt.Errorf("db connect: %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	created, err := sqlcgen.New(conn).CreateUsuarioIfAbsent(ctx, sqlcgen.CreateUsuarioIfAbsentParams{
		Username:     cfg.username,
		PasswordHash: hash,
		Nombre:       cfg.nombre,
		Rol:          "admin",
	})
	if err != nil {
		return fmt.Errorf("creating admin: %w", err)
	}
	if created == 0 {
		fmt.Println("admin already exists; nothing changed")
		return nil
	}
	fmt.Println("admin created")
	return nil
}

func loadSeedConfig(getenv config.Getenv) (seedConfig, error) {
	var errs []error
	databaseURL, err := config.Secret(getenv, "DATABASE_URL")
	if err != nil {
		errs = append(errs, err)
	} else if databaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	password, err := config.Secret(getenv, "SEED_ADMIN_PASSWORD")
	if err != nil {
		errs = append(errs, err)
	} else if utf8.RuneCountInString(password) < minPasswordLength {
		errs = append(errs, fmt.Errorf("SEED_ADMIN_PASSWORD must have at least %d characters", minPasswordLength))
	}
	cfg := seedConfig{
		databaseURL: databaseURL,
		username:    getenv("SEED_ADMIN_USERNAME"),
		password:    password,
		nombre:      getenv("SEED_ADMIN_NOMBRE"),
	}
	if n := utf8.RuneCountInString(cfg.username); n < 3 || n > 50 {
		errs = append(errs, errors.New("SEED_ADMIN_USERNAME must have 3 to 50 characters"))
	}
	if n := utf8.RuneCountInString(cfg.nombre); n < 2 || n > 120 {
		errs = append(errs, errors.New("SEED_ADMIN_NOMBRE must have 2 to 120 characters"))
	}
	return cfg, errors.Join(errs...)
}
