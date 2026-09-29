CREATE TABLE ledger_accounts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id UUID REFERENCES users (id) ON DELETE RESTRICT,
    system_account TEXT CHECK (system_account IN (
        'stripe_clearing', 'crypto_clearing', 'fees', 'treasury', 'mint', 'burn', 'withdrawn', 'unclaimed'
    )),
    item_id INTEGER REFERENCES items (id) ON DELETE RESTRICT,
    bucket TEXT CHECK (bucket IN ('card', 'crypto', 'earned_pending', 'earned')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ledger_accounts_single_owner CHECK ((user_id IS NULL) <> (system_account IS NULL)),
    CONSTRAINT ledger_accounts_bucket_only_for_player_nc CHECK (
        (bucket IS NOT NULL) = (user_id IS NOT NULL AND item_id IS NULL)
    ),
    CONSTRAINT ledger_accounts_owner_asset_bucket_key UNIQUE NULLS NOT DISTINCT (user_id, system_account, item_id, bucket)
);

CREATE TABLE ledger_balances (
    account_id BIGINT PRIMARY KEY REFERENCES ledger_accounts (id) ON DELETE RESTRICT,
    available BIGINT NOT NULL DEFAULT 0,
    held BIGINT NOT NULL DEFAULT 0,
    negative_available_policy TEXT NOT NULL CHECK (negative_available_policy IN ('never', 'dispute_only', 'always')),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ledger_balances_held_not_negative CHECK (held >= 0),
    CONSTRAINT ledger_balances_available_not_negative CHECK (available >= 0 OR negative_available_policy <> 'never')
);

CREATE TABLE ledger_journals (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL CHECK (type IN (
        'ENTRY_FEE', 'TOPUP_CARD', 'TOPUP_CRYPTO', 'SHOP_PURCHASE', 'STARTER_PACK', 'MISSION_FUEL', 'MISSION_LOOT',
        'CRAFT_START', 'CRAFT_OUTPUT', 'UPGRADE', 'TRADE_FILL', 'AUCTION_SETTLE', 'WITHDRAWAL', 'DEPOSIT',
        'EARNED_SETTLE', 'REFUND', 'DISPUTE_DEBIT', 'ADMIN_ADJUSTMENT', 'DEV_CREDIT'
    )),
    ref_type TEXT NOT NULL,
    ref_id TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ledger_journals_business_event_key UNIQUE (type, ref_type, ref_id)
);

CREATE TABLE ledger_entries (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    journal_id UUID NOT NULL REFERENCES ledger_journals (id) ON DELETE RESTRICT,
    account_id BIGINT NOT NULL REFERENCES ledger_accounts (id) ON DELETE RESTRICT,
    amount BIGINT NOT NULL CHECK (amount <> 0),
    synced_onchain BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ledger_entries_journal_idx ON ledger_entries (journal_id);
CREATE INDEX ledger_entries_account_idx ON ledger_entries (account_id, id);

CREATE TABLE ledger_holds (
    id UUID PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES ledger_accounts (id) ON DELETE RESTRICT,
    amount BIGINT NOT NULL CHECK (amount > 0),
    remaining BIGINT NOT NULL,
    ref_type TEXT NOT NULL,
    ref_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'CAPTURED', 'RELEASED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ledger_holds_remaining_in_range CHECK (remaining >= 0 AND remaining <= amount),
    CONSTRAINT ledger_holds_active_means_remaining CHECK ((status = 'ACTIVE') = (remaining > 0)),
    CONSTRAINT ledger_holds_reference_key UNIQUE (account_id, ref_type, ref_id)
);

CREATE INDEX ledger_holds_active_account_idx ON ledger_holds (account_id) WHERE status = 'ACTIVE';
CREATE INDEX ledger_holds_reference_idx ON ledger_holds (ref_type, ref_id);

WITH seeded_accounts AS (
    INSERT INTO ledger_accounts (system_account)
    VALUES ('stripe_clearing'), ('crypto_clearing'), ('fees'), ('treasury'), ('withdrawn'), ('unclaimed')
    RETURNING id, system_account
)
INSERT INTO ledger_balances (account_id, negative_available_policy)
SELECT id, CASE WHEN system_account IN ('stripe_clearing', 'crypto_clearing') THEN 'always' ELSE 'never' END
FROM seeded_accounts;
