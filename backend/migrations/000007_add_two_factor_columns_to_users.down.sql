ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_totp_enabled_requires_secret,
    DROP COLUMN IF EXISTS totp_last_used_step,
    DROP COLUMN IF EXISTS totp_enabled_at;
