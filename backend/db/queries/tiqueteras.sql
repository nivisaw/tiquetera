-- name: CreateTiquetera :one
INSERT INTO tiqueteras (cliente_id, almuerzos_comprados, saldo, fecha_vencimiento, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetActiveTiqueteraByCliente :one
SELECT * FROM tiqueteras WHERE cliente_id = $1 AND activa;

-- name: DeactivateActiveTiquetera :one
UPDATE tiqueteras SET activa = false WHERE cliente_id = $1 AND activa
RETURNING saldo;

-- name: DebitTiquetera :one
UPDATE tiqueteras t
SET saldo = t.saldo - 1
WHERE t.cliente_id = sqlc.arg(cliente_id)
  AND t.activa
  AND t.saldo > 0
  AND t.fecha_vencimiento >= sqlc.arg(hoy)::date
  AND EXISTS (SELECT 1 FROM clientes c WHERE c.id = t.cliente_id AND c.activo)
RETURNING t.*;
