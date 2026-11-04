-- name: ListConsumosByCliente :many
SELECT * FROM consumos
WHERE cliente_id = $1
ORDER BY fecha_hora DESC
LIMIT LEAST(sqlc.arg(size)::int, 50) OFFSET sqlc.arg(page_offset)::int;

-- name: CreateAuditLog :exec
INSERT INTO audit_log (usuario_id, accion, entidad, entidad_id, datos)
VALUES ($1, $2, $3, $4, $5);

-- name: CreateConsumo :one
INSERT INTO consumos (tiquetera_id, cliente_id, cajero_id, saldo_resultante, idempotency_key)
VALUES ($1, $2, $3, $4, $5)
RETURNING fecha_hora;

-- name: GetConsumoResultadoByKey :one
SELECT c.cajero_id, c.saldo_resultante, c.fecha_hora, t.fecha_vencimiento, cl.nombre
FROM consumos c
JOIN tiqueteras t ON t.id = c.tiquetera_id
JOIN clientes cl ON cl.id = c.cliente_id
WHERE c.idempotency_key = $1;

-- name: CreateConsumoRechazado :exec
INSERT INTO consumos_rechazados (cajero_id, cliente_id, motivo)
VALUES ($1, $2, $3);
