-- name: CreateActivationToken :exec
INSERT INTO account_activation_tokens (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4);
