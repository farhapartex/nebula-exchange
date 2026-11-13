ALTER TABLE payments ALTER COLUMN plan_id DROP NOT NULL;
ALTER TABLE payments ADD COLUMN coin_pack_id TEXT REFERENCES coin_packs (id) ON DELETE RESTRICT;
ALTER TABLE payments ADD COLUMN coins BIGINT;
ALTER TABLE payments ADD CONSTRAINT payments_buys_a_plan_or_a_coin_pack CHECK (
    (plan_id IS NOT NULL AND coin_pack_id IS NULL AND coins IS NULL)
    OR (plan_id IS NULL AND coin_pack_id IS NOT NULL AND coins > 0)
);

CREATE INDEX payments_coin_pack_id_index ON payments (coin_pack_id);
