ALTER TABLE chapter_unlocks ADD COLUMN payment_id UUID REFERENCES payments (id) ON DELETE RESTRICT;
ALTER TABLE chapter_unlocks ADD CONSTRAINT chapter_unlocks_purchase_has_payment CHECK ((source = 'PURCHASE') = (payment_id IS NOT NULL));
CREATE INDEX chapter_unlocks_payment_id_index ON chapter_unlocks (payment_id);
