-- name: InsertPayment :one
INSERT INTO payments (id, user_id, purpose, method, amount_micro, sku, quantity, provider_session_id, checkout_url, expires_at, created_at, updated_at)
VALUES (
    @id, @user_id, @purpose, @method, @amount_micro, sqlc.narg(sku)::text, @quantity,
    sqlc.narg(provider_session_id)::text, sqlc.narg(checkout_url)::text, @expires_at, @created_at, @created_at
)
RETURNING *;

-- name: GetPaymentForUser :one
SELECT * FROM payments WHERE id = @id AND user_id = @user_id;

-- name: GetPayment :one
SELECT * FROM payments WHERE id = @id;

-- name: ListPaymentsForUser :many
SELECT * FROM payments
WHERE user_id = @user_id
  AND (sqlc.narg(before_id)::uuid IS NULL OR id < sqlc.narg(before_id)::uuid)
ORDER BY id DESC
LIMIT @row_limit;

-- name: SumRecentCardTopups :one
SELECT COALESCE(SUM(amount_micro), 0)::bigint AS total_micro
FROM payments
WHERE user_id = @user_id
  AND purpose = 'TOPUP'
  AND method = 'card'
  AND created_at >= @since
  AND (status IN ('SUCCEEDED', 'REFUNDED', 'DISPUTED') OR (status = 'PENDING' AND expires_at > @now));

-- name: ClaimPendingPayment :one
UPDATE payments
SET status = 'SUCCEEDED', credited_micro = @credited_micro, succeeded_at = @succeeded_at, updated_at = @succeeded_at
WHERE id = @id AND status = 'PENDING'
RETURNING *;

-- name: RecordPurposeOutcome :exec
UPDATE payments
SET purpose_status = @purpose_status, purpose_failure_code = sqlc.narg(purpose_failure_code)::text, updated_at = now()
WHERE id = @id;

-- name: ExpirePendingPayments :many
UPDATE payments
SET status = 'EXPIRED', updated_at = @now
WHERE status = 'PENDING' AND expires_at <= @now
RETURNING id;

-- name: InsertExternalEvent :one
INSERT INTO external_events (id, provider, event_type, payload)
VALUES (@id, @provider, @event_type, @payload)
ON CONFLICT (id) DO NOTHING
RETURNING id;

-- name: MarkExternalEventProcessed :exec
UPDATE external_events SET processed_at = now() WHERE id = @id;
