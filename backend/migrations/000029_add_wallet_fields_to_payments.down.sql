ALTER TABLE payments
    DROP CONSTRAINT IF EXISTS payments_lowercase_wallet_values,
    DROP CONSTRAINT IF EXISTS payments_wallet_payment_recorded,
    DROP CONSTRAINT IF EXISTS payments_wallet_fields_present,
    DROP CONSTRAINT IF EXISTS payments_transaction_hash_unique,
    DROP CONSTRAINT IF EXISTS payments_payment_reference_unique;

DELETE FROM payments WHERE payment_method = 'WALLET';

ALTER TABLE payments
    DROP COLUMN IF EXISTS wallet_amount_units,
    DROP COLUMN IF EXISTS wallet_asset,
    DROP COLUMN IF EXISTS block_number,
    DROP COLUMN IF EXISTS transaction_hash,
    DROP COLUMN IF EXISTS reported_transaction_hash,
    DROP COLUMN IF EXISTS authorization_signature,
    DROP COLUMN IF EXISTS payer_address,
    DROP COLUMN IF EXISTS payment_reference,
    DROP COLUMN IF EXISTS payment_method;

DROP TYPE IF EXISTS wallet_payment_asset;
DROP TYPE IF EXISTS payment_method;
