CREATE TABLE account_activation_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX account_activation_tokens_user_id_index ON account_activation_tokens (user_id);
CREATE INDEX account_activation_tokens_expires_at_index ON account_activation_tokens (expires_at);
