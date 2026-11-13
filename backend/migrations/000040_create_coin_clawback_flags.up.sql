CREATE TABLE coin_clawback_flags (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    payment_id UUID NOT NULL REFERENCES payments (id) ON DELETE RESTRICT,
    pack_coins BIGINT NOT NULL,
    clawed_back_coins BIGINT NOT NULL,
    missing_coins BIGINT NOT NULL,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT coin_clawback_flags_once_per_payment UNIQUE (payment_id),
    CONSTRAINT coin_clawback_flags_amounts_valid CHECK (
        pack_coins > 0 AND clawed_back_coins >= 0 AND missing_coins > 0
        AND clawed_back_coins + missing_coins = pack_coins
    )
);

CREATE INDEX coin_clawback_flags_open_index ON coin_clawback_flags (created_at DESC) WHERE reviewed_at IS NULL;
CREATE INDEX coin_clawback_flags_user_index ON coin_clawback_flags (user_id);
