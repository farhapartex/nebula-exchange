package ledgertest

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/ledger"
)

func CreatePlayer(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	userID, _ := uuid.NewV7()
	shortID := userID.String()[24:]
	if _, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, email, username, password_hash, status, is_active, activated_at, terms_accepted_at)
		VALUES ($1, $2, $3, 'not-a-real-hash', 'ACTIVE', true, now(), now())`,
		userID, "ledger-"+shortID+"@nebula.test", "pilot_"+shortID,
	); err != nil {
		t.Fatalf("create player: %v", err)
	}
	return userID
}

func InTransaction(t *testing.T, pool *pgxpool.Pool, work func(tx pgx.Tx) error) error {
	t.Helper()
	return pgx.BeginFunc(context.Background(), pool, work)
}

func Fund(t *testing.T, pool *pgxpool.Pool, account ledger.AccountKey, amount int64) {
	t.Helper()
	if err := InTransaction(t, pool, func(tx pgx.Tx) error {
		_, err := ledger.DevCredit(context.Background(), tx, account, amount)
		return err
	}); err != nil {
		t.Fatalf("fund %+v with %d: %v", account, amount, err)
	}
}

type Balance struct {
	Available int64
	Held      int64
}

func BalanceOf(t *testing.T, pool *pgxpool.Pool, account ledger.AccountKey) Balance {
	t.Helper()
	available, held, err := ledger.AccountBalance(context.Background(), pool, account)
	if err != nil {
		t.Fatalf("read balance of %+v: %v", account, err)
	}
	return Balance{Available: available, Held: held}
}

func RequireIntegrity(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	report, err := ledger.RunIntegrityCheck(context.Background(), pool)
	if err != nil {
		t.Fatalf("run integrity check: %v", err)
	}
	if report.MismatchCount() != 0 {
		t.Fatalf("ledger integrity broken: %+v", report)
	}
}

func Hold(t *testing.T, pool *pgxpool.Pool, account ledger.AccountKey, amount int64) uuid.UUID {
	t.Helper()
	var holdID uuid.UUID
	if err := InTransaction(t, pool, func(tx pgx.Tx) error {
		var holdErr error
		holdID, holdErr = ledger.Hold(context.Background(), tx, account, amount, ledger.Reference{Type: "test_hold", ID: uuid.NewString()})
		return holdErr
	}); err != nil {
		t.Fatalf("hold %d on %+v: %v", amount, account, err)
	}
	return holdID
}
