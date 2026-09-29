-- name: DeleteNeverActivatedUsers :execrows
DELETE FROM users
WHERE is_active = false AND created_at < sqlc.arg(created_before)::timestamptz;

-- name: DeleteExpiredActivationTokens :execrows
DELETE FROM account_activation_tokens
WHERE expires_at < sqlc.arg(expired_before)::timestamptz;

-- name: DeleteExpiredPasswordResetTokens :execrows
DELETE FROM password_reset_tokens
WHERE expires_at < sqlc.arg(expired_before)::timestamptz;

-- name: DeleteStaleRefreshTokens :execrows
DELETE FROM refresh_tokens
WHERE expires_at < sqlc.arg(stale_before)::timestamptz
   OR (revoked_at IS NOT NULL AND revoked_at < sqlc.arg(stale_before)::timestamptz);

-- name: DeleteOldIdempotencyKeys :execrows
DELETE FROM idempotency_keys
WHERE created_at < sqlc.arg(created_before)::timestamptz;

-- name: DeleteOldSentEmails :execrows
DELETE FROM email_outbox
WHERE status = 'sent' AND sent_at < sqlc.arg(sent_before)::timestamptz;
