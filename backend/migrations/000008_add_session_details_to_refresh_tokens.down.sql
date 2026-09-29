DROP INDEX IF EXISTS refresh_tokens_active_by_user_index;
ALTER TABLE refresh_tokens
    DROP COLUMN IF EXISTS session_started_at,
    DROP COLUMN IF EXISTS ip_address,
    DROP COLUMN IF EXISTS user_agent;
