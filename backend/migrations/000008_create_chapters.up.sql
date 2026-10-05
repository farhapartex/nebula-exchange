CREATE TABLE chapters (
    id TEXT PRIMARY KEY,
    number INTEGER NOT NULL,
    title TEXT NOT NULL,
    summary TEXT NOT NULL,
    is_free BOOLEAN NOT NULL DEFAULT false,
    price_coins BIGINT,
    is_published BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chapters_number_unique UNIQUE (number),
    CONSTRAINT chapters_number_positive CHECK (number > 0),
    CONSTRAINT chapters_price_matches_access CHECK ((is_free AND price_coins IS NULL) OR (NOT is_free AND price_coins IS NOT NULL AND price_coins > 0))
);
