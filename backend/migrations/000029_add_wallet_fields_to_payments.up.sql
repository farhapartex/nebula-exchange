CREATE TYPE payment_method AS ENUM ('CARD', 'WALLET');
CREATE TYPE wallet_payment_asset AS ENUM ('ETH', 'USDC');

ALTER TABLE payments
    ADD COLUMN payment_method payment_method NOT NULL DEFAULT 'CARD',
    ADD COLUMN payment_reference CHAR(66),
    ADD COLUMN payer_address CHAR(42),
    ADD COLUMN authorization_signature TEXT,
    ADD COLUMN reported_transaction_hash CHAR(66),
    ADD COLUMN transaction_hash CHAR(66),
    ADD COLUMN block_number BIGINT,
    ADD COLUMN wallet_asset wallet_payment_asset,
    ADD COLUMN wallet_amount_units NUMERIC(78, 0);

ALTER TABLE payments
    ADD CONSTRAINT payments_payment_reference_unique UNIQUE (payment_reference),
    ADD CONSTRAINT payments_transaction_hash_unique UNIQUE (transaction_hash),
    ADD CONSTRAINT payments_wallet_fields_present CHECK (
        payment_method = 'CARD'
        OR (
            payment_reference IS NOT NULL
            AND payer_address IS NOT NULL
            AND authorization_signature IS NOT NULL
        )
    ),
    ADD CONSTRAINT payments_wallet_payment_recorded CHECK (
        payment_method = 'CARD'
        OR status NOT IN ('PAID', 'REFUNDED', 'DISPUTED')
        OR (transaction_hash IS NOT NULL AND wallet_asset IS NOT NULL AND wallet_amount_units IS NOT NULL)
    ),
    ADD CONSTRAINT payments_lowercase_wallet_values CHECK (
        payer_address = lower(payer_address)
        AND payment_reference = lower(payment_reference)
        AND transaction_hash = lower(transaction_hash)
        AND reported_transaction_hash = lower(reported_transaction_hash)
    );
