ALTER TABLE users
    ADD COLUMN totp_enabled_at TIMESTAMPTZ,
    ADD COLUMN totp_last_used_step BIGINT,
    ADD CONSTRAINT users_totp_enabled_requires_secret CHECK (totp_enabled_at IS NULL OR totp_secret_enc IS NOT NULL);
