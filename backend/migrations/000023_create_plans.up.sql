CREATE TYPE plan_kind AS ENUM ('SINGLE_CHAPTER', 'CHAPTER_BUNDLE', 'ALL_CHAPTERS');

CREATE TABLE plans (
    id TEXT PRIMARY KEY,
    kind plan_kind NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    discount_tiers JSONB NOT NULL DEFAULT '[]',
    sort_order INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT plans_kind_unique UNIQUE (kind),
    CONSTRAINT plans_sort_order_unique UNIQUE (sort_order),
    CONSTRAINT plans_discount_tiers_is_array CHECK (jsonb_typeof(discount_tiers) = 'array')
);
