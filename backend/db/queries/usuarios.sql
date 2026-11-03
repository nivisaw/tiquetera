-- name: GetUsuarioByID :one
SELECT * FROM usuarios WHERE id = $1;

-- name: GetUsuarioByUsername :one
SELECT * FROM usuarios WHERE username = $1;

-- name: CreateUsuarioIfAbsent :execrows
INSERT INTO usuarios (username, password_hash, nombre, rol)
VALUES ($1, $2, $3, $4)
ON CONFLICT (username) DO NOTHING;
