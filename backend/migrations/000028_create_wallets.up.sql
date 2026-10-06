CREATE TABLE wallets (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    address CHAR(42) NOT NULL CHECK (address = lower(address)),
    chain_id BIGINT NOT NULL,
    linked_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT wallets_user_id_unique UNIQUE (user_id),
    CONSTRAINT wallets_address_unique UNIQUE (address)
);
