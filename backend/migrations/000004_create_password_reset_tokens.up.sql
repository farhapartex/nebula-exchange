CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT password_reset_tokens_token_hash_unique UNIQUE (token_hash),
    CONSTRAINT password_reset_tokens_expiry_after_creation CHECK (expires_at > created_at)
);

CREATE INDEX password_reset_tokens_user_id_index ON password_reset_tokens (user_id);
CREATE INDEX password_reset_tokens_expires_at_index ON password_reset_tokens (expires_at);
