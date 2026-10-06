UPDATE chapters SET price_cents = price_cents * 1000000 WHERE price_cents IS NOT NULL;
ALTER TABLE chapters RENAME COLUMN price_cents TO price_coins;
