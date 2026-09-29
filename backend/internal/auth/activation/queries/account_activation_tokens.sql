-- name: CreateActivationToken :exec
INSERT INTO account_activation_tokens (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: FindActivationTokenWithUser :one
SELECT
    tokens.user_id,
    tokens.expires_at,
    tokens.used_at,
    users.email,
    users.username,
    users.is_active
FROM account_activation_tokens AS tokens
JOIN users ON users.id = tokens.user_id
WHERE tokens.token_hash = $1;

-- name: ConsumeActivationToken :one
UPDATE account_activation_tokens
SET used_at = sqlc.arg(now)::timestamptz
WHERE token_hash = sqlc.arg(token_hash)
  AND used_at IS NULL
  AND expires_at > sqlc.arg(now)::timestamptz
RETURNING user_id;

-- name: ExpireUnusedActivationTokens :exec
UPDATE account_activation_tokens
SET expires_at = sqlc.arg(now)::timestamptz
WHERE user_id = sqlc.arg(user_id)
  AND used_at IS NULL
  AND expires_at > sqlc.arg(now)::timestamptz;
