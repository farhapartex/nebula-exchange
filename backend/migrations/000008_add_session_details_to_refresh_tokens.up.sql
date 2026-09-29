ALTER TABLE refresh_tokens
    ADD COLUMN user_agent TEXT NOT NULL DEFAULT '',
    ADD COLUMN ip_address TEXT NOT NULL DEFAULT '',
    ADD COLUMN session_started_at TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE INDEX refresh_tokens_active_by_user_index ON refresh_tokens (user_id, created_at DESC) WHERE revoked_at IS NULL;
