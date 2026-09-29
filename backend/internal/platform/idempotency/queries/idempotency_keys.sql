-- name: InsertInProgressKey :execrows
INSERT INTO idempotency_keys (scope, idempotency_key, request_hash, status)
VALUES ($1, $2, $3, 'in_progress')
ON CONFLICT (scope, idempotency_key) DO NOTHING;

-- name: FindKey :one
SELECT scope, idempotency_key, request_hash, status, response_status_code, response_content_type, response_body, created_at
FROM idempotency_keys
WHERE scope = $1 AND idempotency_key = $2;

-- name: CompleteKey :exec
UPDATE idempotency_keys
SET status = 'completed',
    response_status_code = $3,
    response_content_type = $4,
    response_body = $5,
    completed_at = now()
WHERE scope = $1 AND idempotency_key = $2;

-- name: DeleteKey :exec
DELETE FROM idempotency_keys
WHERE scope = $1 AND idempotency_key = $2;

-- name: DeleteKeyCreatedBefore :execrows
DELETE FROM idempotency_keys
WHERE scope = $1 AND idempotency_key = $2 AND created_at < $3;
