CREATE TYPE coin_entry_reason AS ENUM ('WELCOME_GIFT', 'COIN_PACK', 'COIN_PACK_CLAWBACK');

CREATE TABLE coin_balances (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE RESTRICT,
    coins BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT coin_balances_not_negative CHECK (coins >= 0)
);

CREATE TABLE coin_ledger_entries (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    amount BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    reason coin_entry_reason NOT NULL,
    payment_id UUID REFERENCES payments (id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT coin_ledger_entries_amount_not_zero CHECK (amount <> 0),
    CONSTRAINT coin_ledger_entries_balance_after_not_negative CHECK (balance_after >= 0),
    CONSTRAINT coin_ledger_entries_payment_matches_reason CHECK (
        (reason = 'WELCOME_GIFT' AND payment_id IS NULL AND amount > 0)
        OR (reason = 'COIN_PACK' AND payment_id IS NOT NULL AND amount > 0)
        OR (reason = 'COIN_PACK_CLAWBACK' AND payment_id IS NOT NULL AND amount < 0)
    )
);

CREATE INDEX coin_ledger_entries_user_index ON coin_ledger_entries (user_id, created_at DESC, id DESC);
CREATE UNIQUE INDEX coin_ledger_entries_one_welcome_gift ON coin_ledger_entries (user_id) WHERE reason = 'WELCOME_GIFT';
CREATE UNIQUE INDEX coin_ledger_entries_once_per_payment ON coin_ledger_entries (payment_id, reason) WHERE payment_id IS NOT NULL;

CREATE FUNCTION reject_coin_ledger_entry_changes() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'coin ledger entries cannot be changed or deleted';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER coin_ledger_entries_append_only
    BEFORE UPDATE OR DELETE ON coin_ledger_entries
    FOR EACH ROW EXECUTE FUNCTION reject_coin_ledger_entry_changes();
