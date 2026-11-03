-- El LOGIN y la contraseña se asignan fuera de las migraciones (make migrate).
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'tiquetera_app') THEN
        CREATE ROLE tiquetera_app NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE;
    END IF;
END;
$$;

GRANT SELECT, INSERT, UPDATE ON usuarios, refresh_tokens, clientes, tiqueteras TO tiquetera_app;
GRANT SELECT, INSERT ON consumos, consumos_rechazados, audit_log TO tiquetera_app;
