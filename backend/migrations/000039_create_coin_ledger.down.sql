DROP TRIGGER IF EXISTS coin_ledger_entries_append_only ON coin_ledger_entries;
DROP FUNCTION IF EXISTS reject_coin_ledger_entry_changes();
DROP TABLE IF EXISTS coin_ledger_entries;
DROP TABLE IF EXISTS coin_balances;
DROP TYPE IF EXISTS coin_entry_reason;
