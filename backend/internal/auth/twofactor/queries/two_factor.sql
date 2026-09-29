-- name: FindTwoFactorState :one
SELECT email, username, totp_secret_enc, totp_enabled_at
FROM users
WHERE id = $1;

-- name: SetPendingTotpSecret :execrows
UPDATE users
SET totp_secret_enc = sqlc.arg(totp_secret_enc), updated_at = sqlc.arg(now)::timestamptz
WHERE id = sqlc.arg(id) AND totp_enabled_at IS NULL;

-- name: EnableTotp :execrows
UPDATE users
SET totp_enabled_at = sqlc.arg(now)::timestamptz, updated_at = sqlc.arg(now)::timestamptz
WHERE id = sqlc.arg(id) AND totp_enabled_at IS NULL AND totp_secret_enc IS NOT NULL;

-- name: DisableTotp :exec
UPDATE users
SET totp_secret_enc = NULL, totp_enabled_at = NULL, totp_last_used_step = NULL, updated_at = sqlc.arg(now)::timestamptz
WHERE id = sqlc.arg(id);

-- name: ClaimTotpStep :execrows
UPDATE users
SET totp_last_used_step = sqlc.arg(step)::bigint
WHERE id = sqlc.arg(id) AND (totp_last_used_step IS NULL OR totp_last_used_step < sqlc.arg(step)::bigint);
