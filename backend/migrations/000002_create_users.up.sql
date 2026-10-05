CREATE TYPE user_status AS ENUM ('UNVERIFIED', 'ACTIVE', 'FROZEN', 'BANNED');

CREATE TABLE users (
    id UUID PRIMARY KEY,
    email CITEXT NOT NULL,
    username CITEXT NOT NULL,
    password_hash TEXT NOT NULL,
    status user_status NOT NULL DEFAULT 'UNVERIFIED',
    is_active BOOLEAN NOT NULL DEFAULT false,
    is_admin BOOLEAN NOT NULL DEFAULT false,
    terms_accepted_at TIMESTAMPTZ NOT NULL,
    activated_at TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_username_unique UNIQUE (username),
    CONSTRAINT users_email_length CHECK (char_length(email) BETWEEN 3 AND 254),
    CONSTRAINT users_username_format CHECK (username ~ '^[A-Za-z0-9_]{3,20}$'),
    CONSTRAINT users_activation_consistency CHECK (is_active = (activated_at IS NOT NULL)),
    CONSTRAINT users_unverified_until_activated CHECK ((status = 'UNVERIFIED') = (NOT is_active))
);

CREATE INDEX users_inactive_created_at_index ON users (created_at) WHERE NOT is_active;
