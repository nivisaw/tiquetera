ALTER TABLE tiqueteras ADD COLUMN almuerzos_vendidos integer;

UPDATE tiqueteras t
SET almuerzos_vendidos = t.almuerzos_comprados - COALESCE((
    SELECT p.saldo FROM tiqueteras p
    WHERE p.cliente_id = t.cliente_id AND p.created_at < t.created_at
    ORDER BY p.created_at DESC
    LIMIT 1
), 0);

ALTER TABLE tiqueteras
    ALTER COLUMN almuerzos_vendidos SET NOT NULL,
    ADD CONSTRAINT tiqueteras_vendidos_check
        CHECK (almuerzos_vendidos > 0 AND almuerzos_vendidos <= almuerzos_comprados);

CREATE INDEX tiqueteras_vigentes_idx ON tiqueteras (fecha_vencimiento) WHERE activa;
CREATE INDEX tiqueteras_created_at_idx ON tiqueteras (created_at);
CREATE INDEX consumos_fecha_idx ON consumos (fecha_hora);
CREATE INDEX consumos_rechazados_fecha_idx ON consumos_rechazados (fecha_hora);
