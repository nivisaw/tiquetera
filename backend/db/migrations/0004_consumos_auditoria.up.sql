CREATE TABLE consumos (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tiquetera_id     uuid NOT NULL,
    cliente_id       uuid NOT NULL REFERENCES clientes (id) ON DELETE RESTRICT,
    cajero_id        uuid NOT NULL REFERENCES usuarios (id) ON DELETE RESTRICT,
    fecha_hora       timestamptz NOT NULL DEFAULT now(),
    saldo_resultante integer NOT NULL CHECK (saldo_resultante >= 0),
    idempotency_key  uuid NOT NULL UNIQUE,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT consumos_tiquetera_cliente_fkey FOREIGN KEY (tiquetera_id, cliente_id)
        REFERENCES tiqueteras (id, cliente_id) ON DELETE RESTRICT
);

CREATE INDEX consumos_cliente_fecha_idx ON consumos (cliente_id, fecha_hora DESC);
CREATE INDEX consumos_cajero_fecha_idx ON consumos (cajero_id, fecha_hora DESC);

CREATE TRIGGER consumos_append_only BEFORE UPDATE OR DELETE ON consumos
    FOR EACH ROW EXECUTE FUNCTION reject_mutation();

CREATE TABLE consumos_rechazados (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    cajero_id  uuid NOT NULL REFERENCES usuarios (id) ON DELETE RESTRICT,
    cliente_id uuid REFERENCES clientes (id) ON DELETE RESTRICT,
    motivo     text NOT NULL CHECK (motivo IN (
        'QR_INVALIDO', 'CLIENTE_NO_EXISTE', 'CLIENTE_INACTIVO',
        'SIN_TIQUETERA', 'TIQUETERA_VENCIDA', 'SIN_SALDO'
    )),
    fecha_hora timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX consumos_rechazados_cajero_fecha_idx ON consumos_rechazados (cajero_id, fecha_hora DESC);

CREATE TABLE audit_log (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id uuid REFERENCES usuarios (id) ON DELETE RESTRICT,
    accion     text NOT NULL CHECK (accion <> ''),
    entidad    text NOT NULL CHECK (entidad <> ''),
    entidad_id uuid,
    datos      jsonb NOT NULL DEFAULT '{}',
    fecha_hora timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX audit_log_fecha_idx ON audit_log (fecha_hora DESC);
