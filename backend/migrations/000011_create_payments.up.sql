CREATE TABLE payments (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    purpose TEXT NOT NULL CHECK (purpose IN ('ENTRY_FEE', 'TOPUP', 'SHOP_PURCHASE')),
    method TEXT NOT NULL CHECK (method IN ('card', 'crypto')),
    status TEXT NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'SUCCEEDED', 'FAILED', 'EXPIRED', 'REFUNDED', 'DISPUTED')),
    amount_micro BIGINT NOT NULL CHECK (amount_micro > 0),
    sku TEXT REFERENCES shop_skus (sku) ON DELETE RESTRICT,
    credited_micro BIGINT CHECK (credited_micro IS NULL OR credited_micro > 0),
    purpose_status TEXT NOT NULL DEFAULT 'PENDING' CHECK (purpose_status IN ('PENDING', 'APPLIED', 'FAILED')),
    purpose_failure_code TEXT,
    provider_session_id TEXT UNIQUE,
    checkout_url TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    succeeded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT payments_sku_only_for_shop CHECK ((sku IS NOT NULL) = (purpose = 'SHOP_PURCHASE')),
    CONSTRAINT payments_settled_has_credit CHECK (status IN ('PENDING', 'FAILED', 'EXPIRED') OR credited_micro IS NOT NULL)
);

CREATE INDEX payments_user_idx ON payments (user_id, id DESC);
CREATE INDEX payments_pending_expiry_idx ON payments (expires_at) WHERE status = 'PENDING';

CREATE TABLE external_events (
    id TEXT PRIMARY KEY,
    provider TEXT NOT NULL CHECK (provider IN ('stripe', 'chain', 'development')),
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ
);
