-- name: CreateUser :one
INSERT INTO users (id, email, username, password_hash, terms_accepted_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, email, username, status, is_active, is_admin, terms_accepted_at, activated_at, last_login_at, created_at, updated_at;

-- name: CheckIdentifiersTaken :one
SELECT
    EXISTS (SELECT 1 FROM users WHERE lower(users.email) = lower(sqlc.arg(email)::text)) AS is_email_taken,
    EXISTS (SELECT 1 FROM users WHERE lower(users.username) = lower(sqlc.arg(username)::text)) AS is_username_taken;

-- name: FindUserCredentialsByEmail :one
SELECT id, email, username, password_hash, status, is_active, is_admin, created_at, last_login_at, totp_enabled_at
FROM users
WHERE lower(email) = lower(sqlc.arg(email)::text);

-- name: FindUserByID :one
SELECT id, email, username, status, is_active, is_admin, created_at, activated_at, last_login_at, totp_enabled_at
FROM users
WHERE id = $1;

-- name: RecordLogin :exec
UPDATE users
SET last_login_at = $2, updated_at = $2
WHERE id = $1;

-- name: ActivateUser :one
UPDATE users
SET is_active = true,
    activated_at = sqlc.arg(activated_at)::timestamptz,
    status = CASE WHEN status = 'UNVERIFIED' THEN 'PENDING_PAYMENT' ELSE status END,
    updated_at = sqlc.arg(activated_at)::timestamptz
WHERE id = sqlc.arg(id) AND is_active = false
RETURNING id, email, username, status, is_active, is_admin, created_at, activated_at, last_login_at;

-- name: UpdatePasswordHash :one
UPDATE users
SET password_hash = sqlc.arg(password_hash), updated_at = sqlc.arg(updated_at)::timestamptz
WHERE id = sqlc.arg(id)
RETURNING email;

-- name: UpdateUsername :one
UPDATE users
SET username = sqlc.arg(username), updated_at = sqlc.arg(updated_at)::timestamptz
WHERE id = sqlc.arg(id)
RETURNING id;

-- name: FindPasswordHashByID :one
SELECT password_hash FROM users WHERE id = $1;
