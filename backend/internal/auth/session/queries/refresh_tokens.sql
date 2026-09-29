-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: ConsumeRefreshToken :one
UPDATE refresh_tokens
SET revoked_at = sqlc.arg(now)::timestamptz
WHERE token_hash = sqlc.arg(token_hash)
  AND revoked_at IS NULL
  AND expires_at > sqlc.arg(now)::timestamptz
RETURNING id, user_id;

-- name: FindRefreshTokenByHash :one
SELECT id, user_id, expires_at, revoked_at, replaced_by
FROM refresh_tokens
WHERE token_hash = $1;

-- name: LinkReplacementRefreshToken :exec
UPDATE refresh_tokens
SET replaced_by = sqlc.arg(replaced_by)
WHERE id = sqlc.arg(id);

-- name: RevokeRefreshTokenByHash :exec
UPDATE refresh_tokens
SET revoked_at = sqlc.arg(now)::timestamptz
WHERE token_hash = sqlc.arg(token_hash) AND revoked_at IS NULL;

-- name: RevokeAllActiveRefreshTokensForUser :execrows
UPDATE refresh_tokens
SET revoked_at = sqlc.arg(now)::timestamptz
WHERE user_id = sqlc.arg(user_id)
  AND revoked_at IS NULL
  AND expires_at > sqlc.arg(now)::timestamptz;
