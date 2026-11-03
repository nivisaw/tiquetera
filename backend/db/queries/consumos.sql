-- name: ListConsumosByCliente :many
SELECT * FROM consumos
WHERE cliente_id = $1
ORDER BY fecha_hora DESC
LIMIT LEAST(sqlc.arg(size)::int, 50) OFFSET sqlc.arg(page_offset)::int;

-- name: CreateAuditLog :exec
INSERT INTO audit_log (usuario_id, accion, entidad, entidad_id, datos)
VALUES ($1, $2, $3, $4, $5);
