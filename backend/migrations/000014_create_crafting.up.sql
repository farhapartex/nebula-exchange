CREATE TABLE craft_jobs (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    recipe_id TEXT NOT NULL REFERENCES recipes (id) ON DELETE RESTRICT,
    quantity INTEGER NOT NULL CHECK (quantity BETWEEN 1 AND 100),
    output_item_id INTEGER NOT NULL REFERENCES items (id) ON DELETE RESTRICT,
    output_quantity INTEGER NOT NULL CHECK (output_quantity > 0),
    fee_micro BIGINT NOT NULL CHECK (fee_micro >= 0),
    inputs JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'CRAFTING' CHECK (status IN ('CRAFTING', 'DELIVERED')),
    started_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    delivered_at TIMESTAMPTZ,
    CONSTRAINT craft_jobs_ends_after_start CHECK (ends_at > started_at),
    CONSTRAINT craft_jobs_delivered_has_time CHECK ((status = 'DELIVERED') = (delivered_at IS NOT NULL))
);

CREATE INDEX craft_jobs_user_idx ON craft_jobs (user_id, id DESC);
CREATE INDEX craft_jobs_due_idx ON craft_jobs (ends_at) WHERE status = 'CRAFTING';
CREATE UNIQUE INDEX craft_jobs_one_active_per_user ON craft_jobs (user_id) WHERE status = 'CRAFTING';

ALTER TABLE payments DROP CONSTRAINT payments_purpose_check;
ALTER TABLE payments ADD CONSTRAINT payments_purpose_check CHECK (purpose IN ('ENTRY_FEE', 'TOPUP', 'SHOP_PURCHASE', 'UPGRADE_PURCHASE'));
ALTER TABLE payments ADD COLUMN upgrade_id TEXT REFERENCES upgrades (id) ON DELETE RESTRICT;
ALTER TABLE payments ADD CONSTRAINT payments_upgrade_only_for_upgrade_purchase CHECK ((upgrade_id IS NOT NULL) = (purpose = 'UPGRADE_PURCHASE'));
