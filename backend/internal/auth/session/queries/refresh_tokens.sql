-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: FindRefreshTokenForUpdate :one
SELECT id, user_id, expires_at, revoked_at
FROM refresh_tokens
WHERE token_hash = $1
FOR UPDATE;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = $2, replaced_by = $3
WHERE id = $1;
