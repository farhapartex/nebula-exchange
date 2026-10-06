ALTER TABLE chapters RENAME COLUMN price_coins TO price_cents;
UPDATE chapters SET price_cents = price_cents / 1000000 WHERE price_cents IS NOT NULL;
