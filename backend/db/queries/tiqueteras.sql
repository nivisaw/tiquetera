-- name: CreateTiquetera :one
INSERT INTO tiqueteras (cliente_id, almuerzos_comprados, saldo, fecha_vencimiento, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetActiveTiqueteraByCliente :one
SELECT * FROM tiqueteras WHERE cliente_id = $1 AND activa;

-- name: DeactivateActiveTiquetera :one
UPDATE tiqueteras SET activa = false WHERE cliente_id = $1 AND activa
RETURNING saldo;
