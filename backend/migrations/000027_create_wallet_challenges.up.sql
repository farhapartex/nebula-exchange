CREATE TABLE wallet_challenges (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    address CHAR(42) NOT NULL CHECK (address = lower(address)),
    chain_id BIGINT NOT NULL,
    nonce TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT wallet_challenges_nonce_unique UNIQUE (nonce)
);

CREATE INDEX wallet_challenges_user_id_index ON wallet_challenges (user_id);
