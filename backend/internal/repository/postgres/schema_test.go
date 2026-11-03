package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // migrate driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"tiquetera/internal/repository/postgres/sqlcgen"
)

const (
	codeCheckViolation  = "23514"
	codeUniqueViolation = "23505"
	codeFKViolation     = "23503"
	codeRestrict        = "23001"
	codeNotNull         = "23502"
	codeDenied          = "42501"
	appPassword         = "app-test-password"
)

type fixture struct {
	owner, app *pgx.Conn
	usuarioID  uuid.UUID
	clienteID  uuid.UUID
	tiquetera  uuid.UUID
}

func TestSchema_MigrateUpDownUp_LeavesExpectedTables(t *testing.T) {
	ctx, ownerURL, _ := startDB(t)
	conn := connect(ctx, t, ownerURL)

	m := newMigrator(t, ownerURL)
	require.NoError(t, m.Up())
	assert.Equal(t, 7, countTables(ctx, t, conn))
	require.NoError(t, m.Down())
	assert.Equal(t, 0, countTables(ctx, t, conn))
	require.NoError(t, m.Up())
	assert.Equal(t, 7, countTables(ctx, t, conn))
	assert.ErrorIs(t, m.Up(), migrate.ErrNoChange)
}

func TestSchema_InvalidWrites_AreRejected(t *testing.T) {
	f := setup(t)
	ins := `INSERT INTO tiqueteras (cliente_id, almuerzos_comprados, saldo, fecha_vencimiento, activa, created_by)
	        VALUES ($1, $2, $3, '2030-01-01', $4, $5)`
	insCliente := `INSERT INTO clientes (nombre, telefono, qr_token_hash, created_by) VALUES ($1, $2, $3, $4)`
	hash := make([]byte, 32)
	hash[0] = 9

	tests := []struct {
		name string
		code string
		sql  string
		args []any
	}{
		{"saldo negativo", codeCheckViolation, `UPDATE tiqueteras SET saldo = -1 WHERE id = $1`, []any{f.tiquetera}},
		{"saldo mayor a comprados", codeCheckViolation, `UPDATE tiqueteras SET saldo = 11 WHERE id = $1`, []any{f.tiquetera}},
		{"segunda tiquetera activa", codeUniqueViolation, ins, []any{f.clienteID, 5, 5, true, f.usuarioID}},
		{"almuerzos en cero", codeCheckViolation, ins, []any{f.clienteID, 0, 0, false, f.usuarioID}},
		{"almuerzos sobre 500", codeCheckViolation, ins, []any{f.clienteID, 501, 501, false, f.usuarioID}},
		{"telefono no E164", codeCheckViolation, insCliente, []any{"Ana", "3001234567", hash, f.usuarioID}},
		{"telefono repetido", codeUniqueViolation, insCliente, []any{"Ana", "+573001234567", hash, f.usuarioID}},
		{"nombre corto", codeCheckViolation, insCliente, []any{"A", "+573001234568", hash, f.usuarioID}},
		{"hash QR invalido", codeCheckViolation, insCliente, []any{"Ana", "+573001234568", []byte{1}, f.usuarioID}},
		{"nombre nulo", codeNotNull, insCliente, []any{nil, "+573001234568", hash, f.usuarioID}},
		{"rol invalido", codeCheckViolation, `INSERT INTO usuarios (username, password_hash, nombre, rol) VALUES ('otro', 'x', 'Otro', 'root')`, nil},
		{"motivo invalido", codeCheckViolation, `INSERT INTO consumos_rechazados (cajero_id, motivo) VALUES ($1, 'OTRO')`, []any{f.usuarioID}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertCode(t, f.app, tt.code, tt.sql, tt.args...)
		})
	}
}

func TestSchema_Consumos_AreAppendOnlyAndIdempotent(t *testing.T) {
	f := setup(t)
	ctx := t.Context()
	key := uuid.New()
	ins := `INSERT INTO consumos (tiquetera_id, cliente_id, cajero_id, saldo_resultante, idempotency_key) VALUES ($1, $2, $3, 9, $4)`
	_, err := f.app.Exec(ctx, ins, f.tiquetera, f.clienteID, f.usuarioID, key)
	require.NoError(t, err)

	assertCode(t, f.app, codeUniqueViolation, ins, f.tiquetera, f.clienteID, f.usuarioID, key)
	assertCode(t, f.app, codeDenied, `UPDATE consumos SET saldo_resultante = 1`)
	assertCode(t, f.app, codeDenied, `DELETE FROM consumos`)
	assertCode(t, f.owner, codeRestrict, `UPDATE consumos SET saldo_resultante = 1`)
	assertCode(t, f.owner, codeRestrict, `DELETE FROM consumos`)
}

func TestSchema_ConsumoWithForeignCliente_IsRejected(t *testing.T) {
	f := setup(t)
	hash := make([]byte, 32)
	hash[0] = 2
	var otroCliente uuid.UUID
	err := f.app.QueryRow(t.Context(),
		`INSERT INTO clientes (nombre, telefono, qr_token_hash, created_by) VALUES ('Luz', '+573009999999', $1, $2) RETURNING id`,
		hash, f.usuarioID).Scan(&otroCliente)
	require.NoError(t, err)

	assertCode(t, f.app, codeFKViolation,
		`INSERT INTO consumos (tiquetera_id, cliente_id, cajero_id, saldo_resultante, idempotency_key) VALUES ($1, $2, $3, 9, $4)`,
		f.tiquetera, otroCliente, f.usuarioID, uuid.New())
}

func TestSchema_AppRole_HasLeastPrivilege(t *testing.T) {
	f := setup(t)
	var super bool
	require.NoError(t, f.app.QueryRow(t.Context(),
		`SELECT rolsuper FROM pg_roles WHERE rolname = current_user`).Scan(&super))
	assert.False(t, super)

	for _, stmt := range []string{
		`DELETE FROM clientes`,
		`DELETE FROM usuarios`,
		`DELETE FROM tiqueteras`,
		`DELETE FROM audit_log`,
		`UPDATE audit_log SET accion = 'x'`,
		`DROP TABLE consumos`,
		`CREATE TABLE intruso (id int)`,
	} {
		assertCode(t, f.app, codeDenied, stmt)
	}
}

func TestSchema_DeleteReferencedUsuario_IsRestricted(t *testing.T) {
	f := setup(t)
	assertCode(t, f.owner, codeRestrict, `DELETE FROM clientes WHERE id = $1`, f.clienteID)
	assertCode(t, f.owner, codeRestrict, `DELETE FROM usuarios WHERE id = $1`, f.usuarioID)
}

func TestSchema_UpdateRow_RefreshesUpdatedAt(t *testing.T) {
	f := setup(t)
	ctx := t.Context()
	var after time.Time
	_, err := f.owner.Exec(ctx, `ALTER TABLE clientes DISABLE TRIGGER clientes_set_updated_at`)
	require.NoError(t, err)
	_, err = f.owner.Exec(ctx, `UPDATE clientes SET updated_at = '2000-01-01' WHERE id = $1`, f.clienteID)
	require.NoError(t, err)
	_, err = f.owner.Exec(ctx, `ALTER TABLE clientes ENABLE TRIGGER clientes_set_updated_at`)
	require.NoError(t, err)
	require.NoError(t, f.app.QueryRow(ctx, `UPDATE clientes SET nombre = 'Ana Maria' WHERE id = $1 RETURNING updated_at`, f.clienteID).Scan(&after))
	assert.Greater(t, after.Year(), 2000)
}

func TestCreateUsuarioIfAbsent_ExistingUsername_KeepsOriginal(t *testing.T) {
	f := setup(t)
	ctx := t.Context()
	q := sqlcgen.New(f.app)
	params := sqlcgen.CreateUsuarioIfAbsentParams{Username: "Admin2", PasswordHash: "h1", Nombre: "Admin", Rol: "admin"}

	created, err := q.CreateUsuarioIfAbsent(ctx, params)
	require.NoError(t, err)
	assert.EqualValues(t, 1, created)

	params.Username, params.PasswordHash = "admin2", "h2"
	created, err = q.CreateUsuarioIfAbsent(ctx, params)
	require.NoError(t, err)
	assert.EqualValues(t, 0, created)

	u, err := q.GetUsuarioByUsername(ctx, "ADMIN2")
	require.NoError(t, err)
	assert.Equal(t, "h1", u.PasswordHash)
}

func TestSearchClientes_SizeOverLimit_IsCappedAt50(t *testing.T) {
	f := setup(t)
	ctx := t.Context()
	for i := range 55 {
		hash := make([]byte, 32)
		hash[0], hash[1] = 7, byte(i)
		_, err := f.owner.Exec(ctx,
			`INSERT INTO clientes (nombre, telefono, qr_token_hash, created_by) VALUES ('Lote', $1, $2, $3)`,
			fmt.Sprintf("+5731000000%02d", i), hash, f.usuarioID)
		require.NoError(t, err)
	}

	rows, err := sqlcgen.New(f.app).SearchClientes(ctx, sqlcgen.SearchClientesParams{Q: "lote", Size: 500})
	require.NoError(t, err)
	assert.Len(t, rows, 50)
}

func setup(t *testing.T) fixture {
	t.Helper()
	ctx, ownerURL, appURL := startDB(t)
	require.NoError(t, newMigrator(t, ownerURL).Up())

	owner := connect(ctx, t, ownerURL)
	_, err := owner.Exec(ctx, fmt.Sprintf(`ALTER ROLE tiquetera_app LOGIN PASSWORD '%s'`, appPassword))
	require.NoError(t, err)

	f := fixture{owner: owner, app: connect(ctx, t, appURL)}
	require.NoError(t, owner.QueryRow(ctx,
		`INSERT INTO usuarios (username, password_hash, nombre, rol) VALUES ('admin', 'x', 'Admin', 'admin') RETURNING id`,
	).Scan(&f.usuarioID))
	require.NoError(t, owner.QueryRow(ctx,
		`INSERT INTO clientes (nombre, telefono, qr_token_hash, created_by) VALUES ('Ana', '+573001234567', $1, $2) RETURNING id`,
		make([]byte, 32), f.usuarioID,
	).Scan(&f.clienteID))
	require.NoError(t, owner.QueryRow(ctx,
		`INSERT INTO tiqueteras (cliente_id, almuerzos_comprados, saldo, fecha_vencimiento, created_by)
		 VALUES ($1, 10, 10, '2030-01-01', $2) RETURNING id`,
		f.clienteID, f.usuarioID,
	).Scan(&f.tiquetera))
	return f
}

func startDB(t *testing.T) (ctx context.Context, ownerURL, appURL string) {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test")
	}
	ctx = t.Context()
	container, err := tcpostgres.Run(context.WithoutCancel(ctx), "postgres:18.6-alpine",
		tcpostgres.WithDatabase("tiquetera"),
		tcpostgres.WithUsername("owner"),
		tcpostgres.WithPassword("owner-test-password"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	ownerURL, err = container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	appURL = fmt.Sprintf("postgres://tiquetera_app:%s@%s:%s/tiquetera?sslmode=disable", appPassword, host, port.Port())
	return ctx, ownerURL, appURL
}

func newMigrator(t *testing.T, ownerURL string) *migrate.Migrate {
	t.Helper()
	source, err := iofs.New(os.DirFS("../../../db"), "migrations")
	require.NoError(t, err)
	m, err := migrate.NewWithSourceInstance("iofs", source, "pgx5"+ownerURL[len("postgres"):])
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = m.Close() })
	return m
}

func connect(ctx context.Context, t *testing.T, url string) *pgx.Conn {
	t.Helper()
	conn, err := pgx.Connect(ctx, url)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close(context.WithoutCancel(ctx)) })
	return conn
}

func countTables(ctx context.Context, t *testing.T, conn *pgx.Conn) int {
	t.Helper()
	var n int
	require.NoError(t, conn.QueryRow(ctx,
		`SELECT count(*) FROM information_schema.tables
		 WHERE table_schema = 'public' AND table_name <> 'schema_migrations'`).Scan(&n))
	return n
}

func assertCode(t *testing.T, conn *pgx.Conn, code, sql string, args ...any) {
	t.Helper()
	_, err := conn.Exec(t.Context(), sql, args...)
	var pgErr *pgconn.PgError
	require.True(t, errors.As(err, &pgErr), "expected a PostgreSQL error, got %v", err)
	assert.Equal(t, code, pgErr.Code)
}
