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

-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (usuario_id, token_hash, family_id, expires_at)
VALUES ($1, $2, $3, $4);

-- name: GetRefreshTokenByHashForUpdate :one
SELECT * FROM refresh_tokens WHERE token_hash = $1 FOR UPDATE;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = $2 WHERE id = $1;

-- name: RevokeRefreshFamily :execrows
UPDATE refresh_tokens SET revoked_at = $2 WHERE family_id = $1 AND revoked_at IS NULL;
