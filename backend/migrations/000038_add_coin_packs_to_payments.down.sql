DELETE FROM payments WHERE coin_pack_id IS NOT NULL;
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_buys_a_plan_or_a_coin_pack;
DROP INDEX IF EXISTS payments_coin_pack_id_index;
ALTER TABLE payments DROP COLUMN IF EXISTS coins;
ALTER TABLE payments DROP COLUMN IF EXISTS coin_pack_id;
ALTER TABLE payments ALTER COLUMN plan_id SET NOT NULL;
