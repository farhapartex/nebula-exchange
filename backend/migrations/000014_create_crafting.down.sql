ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_upgrade_only_for_upgrade_purchase;
ALTER TABLE payments DROP COLUMN IF EXISTS upgrade_id;
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_purpose_check;
ALTER TABLE payments ADD CONSTRAINT payments_purpose_check CHECK (purpose IN ('ENTRY_FEE', 'TOPUP', 'SHOP_PURCHASE'));
DROP TABLE IF EXISTS craft_jobs;
