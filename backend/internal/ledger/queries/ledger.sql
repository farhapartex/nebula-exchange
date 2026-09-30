-- name: FindAccount :one
SELECT id
FROM ledger_accounts
WHERE user_id IS NOT DISTINCT FROM sqlc.narg(user_id)::uuid
  AND system_account IS NOT DISTINCT FROM sqlc.narg(system_account)::text
  AND item_id IS NOT DISTINCT FROM sqlc.narg(item_id)::integer
  AND bucket IS NOT DISTINCT FROM sqlc.narg(bucket)::text;

-- name: InsertAccount :one
WITH inserted_account AS (
    INSERT INTO ledger_accounts (user_id, system_account, item_id, bucket)
    VALUES (sqlc.narg(user_id)::uuid, sqlc.narg(system_account)::text, sqlc.narg(item_id)::integer, sqlc.narg(bucket)::text)
    ON CONFLICT ON CONSTRAINT ledger_accounts_owner_asset_bucket_key DO NOTHING
    RETURNING id
), inserted_balance AS (
    INSERT INTO ledger_balances (account_id, negative_available_policy)
    SELECT id, sqlc.arg(negative_available_policy)::text FROM inserted_account
)
SELECT id FROM inserted_account;

-- name: ListAccountsByID :many
SELECT a.id, a.user_id, a.system_account, a.item_id, a.bucket, b.negative_available_policy
FROM ledger_accounts a
JOIN ledger_balances b ON b.account_id = a.id
WHERE a.id = ANY(sqlc.arg(account_ids)::bigint[]);

-- name: InsertJournal :one
INSERT INTO ledger_journals (id, type, ref_type, ref_id, metadata)
VALUES (@id, @type, @ref_type, @ref_id, @metadata)
ON CONFLICT ON CONSTRAINT ledger_journals_business_event_key DO NOTHING
RETURNING id;

-- name: InsertEntries :copyfrom
INSERT INTO ledger_entries (journal_id, account_id, amount) VALUES (@journal_id, @account_id, @amount);

-- name: ApplyAvailableChange :one
UPDATE ledger_balances
SET available = available + @change, updated_at = now()
WHERE account_id = @account_id
  AND (
      available + @change >= 0
      OR negative_available_policy = 'always'
      OR (negative_available_policy = 'dispute_only' AND sqlc.arg(allows_dispute_overdraft)::boolean)
  )
RETURNING available;

-- name: MoveAvailableToHeld :one
UPDATE ledger_balances
SET available = available - @amount, held = held + @amount, updated_at = now()
WHERE account_id = @account_id AND available >= @amount
RETURNING available;

-- name: MoveHeldToAvailable :one
UPDATE ledger_balances
SET available = available + @amount, held = held - @amount, updated_at = now()
WHERE account_id = @account_id AND held >= @amount
RETURNING held;

-- name: TakeFromHeld :one
UPDATE ledger_balances
SET held = held - @amount, updated_at = now()
WHERE account_id = @account_id AND held >= @amount
RETURNING held;

-- name: InsertHold :one
INSERT INTO ledger_holds (id, account_id, amount, remaining, ref_type, ref_id)
VALUES (@id, @account_id, @amount, @amount, @ref_type, @ref_id)
ON CONFLICT ON CONSTRAINT ledger_holds_reference_key DO NOTHING
RETURNING id;

-- name: ReduceHold :one
UPDATE ledger_holds
SET remaining = remaining - @amount,
    status = CASE WHEN remaining - @amount = 0 THEN sqlc.arg(final_status)::text ELSE status END,
    updated_at = now()
WHERE id = @id AND status = 'ACTIVE' AND remaining >= @amount
RETURNING account_id, remaining;

-- name: GetHold :one
SELECT id, account_id, amount, remaining, ref_type, ref_id, status
FROM ledger_holds
WHERE id = @id;

-- name: GetBalance :one
SELECT available, held
FROM ledger_balances
WHERE account_id = @account_id;

-- name: ListPlayerNCBalances :many
SELECT a.id, a.bucket, b.available, b.held
FROM ledger_accounts a
JOIN ledger_balances b ON b.account_id = a.id
WHERE a.user_id = @user_id AND a.item_id IS NULL;

-- name: FindUnbalancedJournals :many
SELECT e.journal_id, COALESCE(a.item_id, 0)::integer AS asset_item_id, SUM(e.amount)::bigint AS total
FROM ledger_entries e
JOIN ledger_accounts a ON a.id = e.account_id
GROUP BY e.journal_id, COALESCE(a.item_id, 0)
HAVING SUM(e.amount) <> 0
LIMIT 50;

-- name: FindBalancesNotMatchingEntries :many
SELECT b.account_id, (b.available + b.held)::bigint AS balance_total, COALESCE(entry_totals.total, 0)::bigint AS entries_total
FROM ledger_balances b
LEFT JOIN (
    SELECT account_id, SUM(amount) AS total FROM ledger_entries GROUP BY account_id
) entry_totals ON entry_totals.account_id = b.account_id
WHERE b.available + b.held <> COALESCE(entry_totals.total, 0)
LIMIT 50;

-- name: FindHeldNotMatchingHolds :many
SELECT b.account_id, b.held, COALESCE(hold_totals.total, 0)::bigint AS active_holds_total
FROM ledger_balances b
LEFT JOIN (
    SELECT account_id, SUM(remaining) AS total FROM ledger_holds WHERE status = 'ACTIVE' GROUP BY account_id
) hold_totals ON hold_totals.account_id = b.account_id
WHERE b.held <> COALESCE(hold_totals.total, 0)
LIMIT 50;

-- name: ListPlayerInventory :many
SELECT a.item_id::integer AS item_id, b.available, b.held
FROM ledger_accounts a
JOIN ledger_balances b ON b.account_id = a.id
WHERE a.user_id = @user_id
  AND a.item_id IS NOT NULL
  AND (b.available <> 0 OR b.held <> 0)
  AND a.item_id > @after_item_id::integer
ORDER BY a.item_id
LIMIT @row_limit;

-- name: ListPlayerJournals :many
SELECT DISTINCT j.id, j.type, j.ref_type, j.ref_id, j.created_at
FROM ledger_accounts a
JOIN ledger_entries e ON e.account_id = a.id
JOIN ledger_journals j ON j.id = e.journal_id
WHERE a.user_id = @user_id
  AND (sqlc.narg(before_journal_id)::uuid IS NULL OR j.id < sqlc.narg(before_journal_id)::uuid)
  AND (sqlc.narg(journal_type)::text IS NULL OR j.type = sqlc.narg(journal_type)::text)
ORDER BY j.id DESC
LIMIT @row_limit;

-- name: ListPlayerJournalEntries :many
SELECT e.journal_id, a.item_id, a.bucket, SUM(e.amount)::bigint AS amount
FROM ledger_entries e
JOIN ledger_accounts a ON a.id = e.account_id
WHERE e.journal_id = ANY(@journal_ids::uuid[]) AND a.user_id = @user_id
GROUP BY e.journal_id, a.item_id, a.bucket
HAVING SUM(e.amount) <> 0
ORDER BY e.journal_id, a.item_id NULLS FIRST, a.bucket;

-- name: ListPlayerJournalTypesAmong :many
SELECT DISTINCT j.type
FROM ledger_accounts a
JOIN ledger_entries e ON e.account_id = a.id
JOIN ledger_journals j ON j.id = e.journal_id
WHERE a.user_id = @user_id AND j.type = ANY(@journal_types::text[]);
