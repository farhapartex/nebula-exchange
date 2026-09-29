-- name: CreatePasswordResetToken :exec
INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: ExpireUnusedPasswordResetTokens :exec
UPDATE password_reset_tokens
SET expires_at = sqlc.arg(now)::timestamptz
WHERE user_id = sqlc.arg(user_id)
  AND used_at IS NULL
  AND expires_at > sqlc.arg(now)::timestamptz;

-- name: FindUsablePasswordResetToken :one
SELECT tokens.user_id, users.email
FROM password_reset_tokens AS tokens
JOIN users ON users.id = tokens.user_id
WHERE tokens.token_hash = sqlc.arg(token_hash)
  AND tokens.used_at IS NULL
  AND tokens.expires_at > sqlc.arg(now)::timestamptz
  AND users.is_active = true;

-- name: ConsumePasswordResetToken :one
UPDATE password_reset_tokens
SET used_at = sqlc.arg(now)::timestamptz
WHERE token_hash = sqlc.arg(token_hash)
  AND used_at IS NULL
  AND expires_at > sqlc.arg(now)::timestamptz
RETURNING user_id;
