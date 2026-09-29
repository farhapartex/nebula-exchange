-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, user_agent, ip_address, session_started_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ConsumeRefreshToken :one
UPDATE refresh_tokens
SET revoked_at = sqlc.arg(now)::timestamptz
WHERE token_hash = sqlc.arg(token_hash)
  AND revoked_at IS NULL
  AND expires_at > sqlc.arg(now)::timestamptz
RETURNING id, user_id, session_started_at;

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

-- name: ListActiveSessions :many
SELECT id, user_agent, ip_address, session_started_at, created_at
FROM refresh_tokens
WHERE user_id = sqlc.arg(user_id)
  AND revoked_at IS NULL
  AND expires_at > sqlc.arg(now)::timestamptz
  AND (sqlc.narg(cursor_created_at)::timestamptz IS NULL
       OR (created_at, id) < (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid))
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(row_limit)::integer;

-- name: RevokeSessionForUser :one
UPDATE refresh_tokens
SET revoked_at = sqlc.arg(now)::timestamptz
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id) AND revoked_at IS NULL
RETURNING token_hash;

-- name: RevokeOtherSessionsForUser :execrows
UPDATE refresh_tokens
SET revoked_at = sqlc.arg(now)::timestamptz
WHERE user_id = sqlc.arg(user_id)
  AND revoked_at IS NULL
  AND token_hash <> sqlc.arg(kept_token_hash);
