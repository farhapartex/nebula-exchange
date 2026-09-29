CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'UNVERIFIED'
        CHECK (status IN ('UNVERIFIED', 'PENDING_PAYMENT', 'ACTIVE', 'FROZEN', 'BANNED')),
    is_active BOOLEAN NOT NULL DEFAULT false,
    is_admin BOOLEAN NOT NULL DEFAULT false,
    totp_secret_enc BYTEA,
    terms_accepted_at TIMESTAMPTZ NOT NULL,
    activated_at TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_activation_consistency CHECK (is_active = (activated_at IS NOT NULL))
);

CREATE UNIQUE INDEX users_email_unique ON users (lower(email));
CREATE UNIQUE INDEX users_username_unique ON users (lower(username));
CREATE INDEX users_inactive_created_at_index ON users (created_at) WHERE is_active = false;
