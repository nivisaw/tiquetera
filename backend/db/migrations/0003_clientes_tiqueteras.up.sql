CREATE TABLE clientes (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre        text NOT NULL CHECK (char_length(nombre) BETWEEN 2 AND 120),
    telefono      text NOT NULL UNIQUE CHECK (telefono ~ '^\+[1-9][0-9]{7,14}$'),
    documento     text CHECK (char_length(documento) BETWEEN 1 AND 30),
    email         citext CHECK (char_length(email::text) BETWEEN 3 AND 254),
    qr_token_hash bytea NOT NULL UNIQUE CHECK (octet_length(qr_token_hash) = 32),
    qr_version    integer NOT NULL DEFAULT 1 CHECK (qr_version >= 1),
    activo        boolean NOT NULL DEFAULT true,
    created_by    uuid NOT NULL REFERENCES usuarios (id) ON DELETE RESTRICT,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER clientes_set_updated_at BEFORE UPDATE ON clientes
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE tiqueteras (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    cliente_id          uuid NOT NULL REFERENCES clientes (id) ON DELETE RESTRICT,
    almuerzos_comprados integer NOT NULL CHECK (almuerzos_comprados > 0 AND almuerzos_comprados <= 500),
    saldo               integer NOT NULL CHECK (saldo >= 0),
    fecha_vencimiento   date NOT NULL,
    activa              boolean NOT NULL DEFAULT true,
    created_by          uuid NOT NULL REFERENCES usuarios (id) ON DELETE RESTRICT,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT tiqueteras_saldo_max CHECK (saldo <= almuerzos_comprados),
    CONSTRAINT tiqueteras_id_cliente_key UNIQUE (id, cliente_id)
);

CREATE UNIQUE INDEX tiqueteras_una_activa_idx ON tiqueteras (cliente_id) WHERE activa;
CREATE INDEX tiqueteras_cliente_id_idx ON tiqueteras (cliente_id);

CREATE TRIGGER tiqueteras_set_updated_at BEFORE UPDATE ON tiqueteras
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
