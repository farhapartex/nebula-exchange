CREATE TABLE coin_packs (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    coins BIGINT NOT NULL,
    price_cents BIGINT NOT NULL,
    sort_order INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT coin_packs_sort_order_unique UNIQUE (sort_order),
    CONSTRAINT coin_packs_amounts_positive CHECK (coins > 0 AND price_cents > 0)
);
