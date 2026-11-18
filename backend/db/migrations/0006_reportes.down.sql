DROP INDEX consumos_rechazados_fecha_idx;
DROP INDEX consumos_fecha_idx;
DROP INDEX tiqueteras_created_at_idx;
DROP INDEX tiqueteras_vigentes_idx;
ALTER TABLE tiqueteras DROP COLUMN almuerzos_vendidos;
