-- name: CreateUser :one
INSERT INTO users (id, email, username, password_hash, terms_accepted_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, email, username, status, is_active, is_admin, terms_accepted_at, activated_at, last_login_at, created_at, updated_at;

-- name: CheckIdentifiersTaken :one
SELECT
    EXISTS (SELECT 1 FROM users WHERE lower(users.email) = lower(sqlc.arg(email)::text)) AS is_email_taken,
    EXISTS (SELECT 1 FROM users WHERE lower(users.username) = lower(sqlc.arg(username)::text)) AS is_username_taken;
