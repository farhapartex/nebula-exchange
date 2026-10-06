DROP INDEX IF EXISTS chapter_unlocks_payment_id_index;
ALTER TABLE chapter_unlocks DROP CONSTRAINT IF EXISTS chapter_unlocks_purchase_has_payment;
ALTER TABLE chapter_unlocks DROP COLUMN IF EXISTS payment_id;
