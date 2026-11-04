-- name: GetUsuarioByID :one
SELECT * FROM usuarios WHERE id = $1;

-- name: GetUsuarioByUsername :one
SELECT * FROM usuarios WHERE username = $1;

-- name: CreateUsuarioIfAbsent :execrows
INSERT INTO usuarios (username, password_hash, nombre, rol)
VALUES ($1, $2, $3, $4)
ON CONFLICT (username) DO NOTHING;

-- name: GetUsuarioByIDForUpdate :one
SELECT * FROM usuarios WHERE id = $1 FOR UPDATE;

-- name: UpdateUsuarioLoginState :exec
UPDATE usuarios SET failed_logins = $2, locked_until = $3 WHERE id = $1;

-- name: CreateUsuario :one
INSERT INTO usuarios (username, password_hash, nombre, rol)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListUsuarios :many
SELECT * FROM usuarios
ORDER BY username, id
LIMIT LEAST(sqlc.arg(size)::int, 50) OFFSET sqlc.arg(page_offset)::int;

-- name: CountUsuarios :one
SELECT count(*) FROM usuarios;

-- name: LockActiveAdmins :many
SELECT id FROM usuarios WHERE rol = 'admin' AND activo ORDER BY id FOR UPDATE;

-- name: UpdateUsuario :one
UPDATE usuarios SET nombre = $2, rol = $3, activo = $4 WHERE id = $1
RETURNING *;

-- name: UpdateUsuarioPassword :execrows
UPDATE usuarios SET password_hash = $2, failed_logins = 0, locked_until = NULL WHERE id = $1;

-- name: RevokeRefreshTokensByUsuario :exec
UPDATE refresh_tokens SET revoked_at = $2 WHERE usuario_id = $1 AND revoked_at IS NULL;

-- name: GetRefreshTokenOwner :one
SELECT usuario_id FROM refresh_tokens WHERE token_hash = $1;

-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (usuario_id, token_hash, family_id, expires_at)
VALUES ($1, $2, $3, $4);

-- name: GetRefreshTokenByHashForUpdate :one
SELECT * FROM refresh_tokens WHERE token_hash = $1 FOR UPDATE;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = $2 WHERE id = $1;

-- name: RevokeRefreshFamily :execrows
UPDATE refresh_tokens SET revoked_at = $2 WHERE family_id = $1 AND revoked_at IS NULL;
