-- name: CreateCliente :one
INSERT INTO clientes (nombre, telefono, documento, email, qr_token_hash, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetClienteByID :one
SELECT * FROM clientes WHERE id = $1;

-- name: GetClienteByQRHash :one
SELECT * FROM clientes WHERE qr_token_hash = $1;

-- name: SearchClientes :many
SELECT * FROM clientes
WHERE nombre ILIKE '%' || sqlc.arg(q)::text || '%'
   OR telefono LIKE '%' || sqlc.arg(q)::text || '%'
ORDER BY nombre, id
LIMIT LEAST(sqlc.arg(size)::int, 50) OFFSET sqlc.arg(page_offset)::int;
