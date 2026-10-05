CREATE TABLE refresh_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    family_id UUID NOT NULL,
    token_hash BYTEA NOT NULL,
    replaced_by_id UUID REFERENCES refresh_sessions (id) ON DELETE SET NULL,
    user_agent TEXT NOT NULL DEFAULT '',
    ip_address INET,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT refresh_sessions_token_hash_unique UNIQUE (token_hash),
    CONSTRAINT refresh_sessions_expiry_after_creation CHECK (expires_at > created_at),
    CONSTRAINT refresh_sessions_not_replaced_by_itself CHECK (replaced_by_id IS NULL OR replaced_by_id <> id)
);

CREATE INDEX refresh_sessions_family_id_index ON refresh_sessions (family_id);
CREATE INDEX refresh_sessions_expires_at_index ON refresh_sessions (expires_at);
CREATE INDEX refresh_sessions_active_by_user_index ON refresh_sessions (user_id, created_at DESC) WHERE revoked_at IS NULL;
