-- name: EnqueueEmail :exec
INSERT INTO email_outbox (id, template, recipient_email, recipient_name, subject, html_body, text_body, max_attempts)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: ClaimDueEmails :many
UPDATE email_outbox
SET claimed_until = sqlc.arg(claimed_until)::timestamptz, updated_at = sqlc.arg(now)::timestamptz
WHERE id IN (
    SELECT candidate.id
    FROM email_outbox AS candidate
    WHERE candidate.status = 'pending'
      AND candidate.next_attempt_at <= sqlc.arg(now)::timestamptz
      AND (candidate.claimed_until IS NULL OR candidate.claimed_until < sqlc.arg(now)::timestamptz)
    ORDER BY candidate.next_attempt_at
    LIMIT sqlc.arg(batch_size)::integer
)
  AND status = 'pending'
  AND (claimed_until IS NULL OR claimed_until < sqlc.arg(now)::timestamptz)
RETURNING id, template, recipient_email, recipient_name, subject, html_body, text_body, attempts, max_attempts;

-- name: MarkEmailSent :exec
UPDATE email_outbox
SET status = 'sent', attempts = attempts + 1, sent_at = sqlc.arg(now)::timestamptz,
    claimed_until = NULL, last_error = NULL, updated_at = sqlc.arg(now)::timestamptz
WHERE id = sqlc.arg(id);

-- name: RecordEmailFailure :exec
UPDATE email_outbox
SET attempts = attempts + 1,
    status = CASE WHEN attempts + 1 >= max_attempts THEN 'failed' ELSE 'pending' END,
    next_attempt_at = sqlc.arg(next_attempt_at)::timestamptz,
    last_error = sqlc.arg(last_error)::text,
    claimed_until = NULL,
    updated_at = sqlc.arg(now)::timestamptz
WHERE id = sqlc.arg(id);
