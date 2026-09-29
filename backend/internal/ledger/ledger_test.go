package ledger_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/ledger/ledgertest"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/database/databasetest"
)

const (
	scoutItemID    = 301
	fuelCellItemID = 401
	oneNC          = int64(1_000_000)
)

func post(t *testing.T, pool *pgxpool.Pool, journal ledger.Journal, captures ...ledger.HoldCapture) error {
	t.Helper()
	return ledgertest.InTransaction(t, pool, func(tx pgx.Tx) error {
		_, err := ledger.Post(context.Background(), tx, journal, captures...)
		return err
	})
}

func hold(t *testing.T, pool *pgxpool.Pool, account ledger.AccountKey, amount int64, referenceID string) (uuid.UUID, error) {
	t.Helper()
	var holdID uuid.UUID
	err := ledgertest.InTransaction(t, pool, func(tx pgx.Tx) error {
		var holdErr error
		holdID, holdErr = ledger.Hold(context.Background(), tx, account, amount, ledger.Reference{Type: "order", ID: referenceID})
		return holdErr
	})
	return holdID, err
}

func TestPostMovesValueAndRejectsDoublePosting(t *testing.T) {
	pool := databasetest.NewPool(t)
	buyer := ledgertest.CreatePlayer(t, pool)
	ledgertest.Fund(t, pool, ledger.PlayerNC(buyer, ledger.BucketCard), 5*oneNC)

	shopPurchase := ledger.Journal{
		Type:      ledger.JournalShopPurchase,
		Reference: ledger.Reference{Type: "payment", ID: "pay-1"},
		Legs: []ledger.Leg{
			ledger.Debit(ledger.PlayerNC(buyer, ledger.BucketCard), 2*oneNC),
			ledger.Credit(ledger.System(ledger.SystemTreasury, ledger.NC), 2*oneNC),
			ledger.Debit(ledger.System(ledger.SystemMint, ledger.Item(scoutItemID)), 1),
			ledger.Credit(ledger.PlayerItem(buyer, scoutItemID), 1),
		},
	}
	if err := post(t, pool, shopPurchase); err != nil {
		t.Fatalf("post: %v", err)
	}
	if err := post(t, pool, shopPurchase); !errors.Is(err, ledger.ErrDuplicateJournal) {
		t.Fatalf("second post of the same event: got %v", err)
	}

	if balance := ledgertest.BalanceOf(t, pool, ledger.PlayerNC(buyer, ledger.BucketCard)); balance.Available != 3*oneNC {
		t.Fatalf("card balance %d", balance.Available)
	}
	if balance := ledgertest.BalanceOf(t, pool, ledger.PlayerItem(buyer, scoutItemID)); balance.Available != 1 {
		t.Fatalf("scouts %d", balance.Available)
	}
	if balance := ledgertest.BalanceOf(t, pool, ledger.System(ledger.SystemMint, ledger.Item(scoutItemID))); balance.Available != -1 {
		t.Fatalf("mint should go negative by design, got %d", balance.Available)
	}
	ledgertest.RequireIntegrity(t, pool)
}

func TestPostRejectsUnbalancedAndOverdrawnJournals(t *testing.T) {
	pool := databasetest.NewPool(t)
	player := ledgertest.CreatePlayer(t, pool)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketEarned), oneNC)

	unbalanced := ledger.Journal{
		Type:      ledger.JournalRefund,
		Reference: ledger.Reference{Type: "test", ID: "unbalanced"},
		Legs: []ledger.Leg{
			ledger.Debit(ledger.PlayerNC(player, ledger.BucketEarned), 10),
			ledger.Credit(ledger.System(ledger.SystemFees, ledger.NC), 9),
		},
	}
	if err := post(t, pool, unbalanced); !errors.Is(err, ledger.ErrUnbalancedJournal) {
		t.Fatalf("unbalanced journal: got %v", err)
	}

	mixedAssets := ledger.Journal{
		Type:      ledger.JournalRefund,
		Reference: ledger.Reference{Type: "test", ID: "mixed"},
		Legs: []ledger.Leg{
			ledger.Debit(ledger.PlayerNC(player, ledger.BucketEarned), 1),
			ledger.Credit(ledger.PlayerItem(player, fuelCellItemID), 1),
		},
	}
	if err := post(t, pool, mixedAssets); !errors.Is(err, ledger.ErrUnbalancedJournal) {
		t.Fatalf("NC must not balance against items: got %v", err)
	}

	overdraw := ledger.Journal{
		Type:      ledger.JournalRefund,
		Reference: ledger.Reference{Type: "test", ID: "overdraw"},
		Legs: []ledger.Leg{
			ledger.Debit(ledger.PlayerNC(player, ledger.BucketEarned), 2*oneNC),
			ledger.Credit(ledger.System(ledger.SystemFees, ledger.NC), 2*oneNC),
		},
	}
	err := post(t, pool, overdraw)
	if !ledger.IsInsufficientBalance(err) {
		t.Fatalf("overdraw: got %v", err)
	}
	if apierror.From(err).Code != apierror.CodeInsufficientFunds {
		t.Fatalf("overdraw maps to %s", apierror.From(err).Code)
	}
	if balance := ledgertest.BalanceOf(t, pool, ledger.PlayerNC(player, ledger.BucketEarned)); balance.Available != oneNC {
		t.Fatalf("failed journals must leave balances untouched, got %d", balance.Available)
	}
	ledgertest.RequireIntegrity(t, pool)
}

func TestOnlyDisputeDebitsMayPushCardNegative(t *testing.T) {
	pool := databasetest.NewPool(t)
	player := ledgertest.CreatePlayer(t, pool)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketCard), oneNC)

	debitCard := func(journalType ledger.JournalType, referenceID string) error {
		return post(t, pool, ledger.Journal{
			Type:      journalType,
			Reference: ledger.Reference{Type: "payment", ID: referenceID},
			Legs: []ledger.Leg{
				ledger.Debit(ledger.PlayerNC(player, ledger.BucketCard), 3*oneNC),
				ledger.Credit(ledger.System(ledger.SystemStripeClearing, ledger.NC), 3*oneNC),
			},
		})
	}
	if err := debitCard(ledger.JournalRefund, "refund"); !ledger.IsInsufficientBalance(err) {
		t.Fatalf("a refund may not overdraw card: %v", err)
	}
	if err := debitCard(ledger.JournalDisputeDebit, "dispute"); err != nil {
		t.Fatalf("a dispute debit may overdraw card: %v", err)
	}
	if balance := ledgertest.BalanceOf(t, pool, ledger.PlayerNC(player, ledger.BucketCard)); balance.Available != -2*oneNC {
		t.Fatalf("card after dispute %d", balance.Available)
	}

	if _, err := pool.Exec(context.Background(), `
		UPDATE ledger_balances b SET available = -1
		FROM ledger_accounts a
		WHERE a.id = b.account_id AND a.system_account = 'fees'`); err == nil {
		t.Fatal("the database itself must reject negative balances on ordinary accounts")
	}
	ledgertest.RequireIntegrity(t, pool)
}

func TestHoldReleaseAndCaptureWithPartialAmounts(t *testing.T) {
	pool := databasetest.NewPool(t)
	buyer := ledgertest.CreatePlayer(t, pool)
	seller := ledgertest.CreatePlayer(t, pool)
	buyerCard := ledger.PlayerNC(buyer, ledger.BucketCard)
	sellerFuel := ledger.PlayerItem(seller, fuelCellItemID)
	ledgertest.Fund(t, pool, buyerCard, 10*oneNC)
	ledgertest.Fund(t, pool, sellerFuel, 20)

	buyerHold, err := hold(t, pool, buyerCard, 6*oneNC, "buy-order")
	if err != nil {
		t.Fatalf("hold NC: %v", err)
	}
	if _, err := hold(t, pool, buyerCard, oneNC, "buy-order"); !errors.Is(err, ledger.ErrDuplicateHold) {
		t.Fatalf("second hold with the same reference: %v", err)
	}
	if _, err := hold(t, pool, buyerCard, 5*oneNC, "too-big"); !ledger.IsInsufficientBalance(err) {
		t.Fatalf("hold beyond available: %v", err)
	}
	sellerHold, err := hold(t, pool, sellerFuel, 20, "sell-order")
	if err != nil {
		t.Fatalf("hold items: %v", err)
	}

	tradeFill := ledger.Journal{
		Type:      ledger.JournalTradeFill,
		Reference: ledger.Reference{Type: "trade", ID: "fill-1"},
		Legs: []ledger.Leg{
			ledger.Credit(ledger.PlayerNC(seller, ledger.BucketEarnedPending), 3_900_000),
			ledger.Credit(ledger.System(ledger.SystemFees, ledger.NC), 100_000),
			ledger.Credit(ledger.PlayerItem(buyer, fuelCellItemID), 8),
		},
	}
	if err := post(t, pool, tradeFill,
		ledger.HoldCapture{HoldID: buyerHold, Amount: 4 * oneNC},
		ledger.HoldCapture{HoldID: sellerHold, Amount: 8},
	); err != nil {
		t.Fatalf("capture fill: %v", err)
	}

	if balance := ledgertest.BalanceOf(t, pool, buyerCard); balance != (ledgertest.Balance{Available: 4 * oneNC, Held: 2 * oneNC}) {
		t.Fatalf("buyer card after fill %+v", balance)
	}
	if balance := ledgertest.BalanceOf(t, pool, sellerFuel); balance != (ledgertest.Balance{Available: 0, Held: 12}) {
		t.Fatalf("seller fuel after fill %+v", balance)
	}

	overCapture := tradeFill
	overCapture.Reference = ledger.Reference{Type: "trade", ID: "fill-2"}
	overCapture.Legs = []ledger.Leg{ledger.Credit(ledger.PlayerNC(seller, ledger.BucketEarnedPending), 3*oneNC)}
	if err := post(t, pool, overCapture, ledger.HoldCapture{HoldID: buyerHold, Amount: 3 * oneNC}); !errors.Is(err, ledger.ErrHoldNotAvailable) {
		t.Fatalf("capturing more than remains: %v", err)
	}

	if err := ledgertest.InTransaction(t, pool, func(tx pgx.Tx) error {
		if err := ledger.Release(context.Background(), tx, buyerHold, oneNC); err != nil {
			return err
		}
		released, err := ledger.ReleaseRemaining(context.Background(), tx, buyerHold)
		if released != oneNC {
			t.Errorf("released remaining %d", released)
		}
		return err
	}); err != nil {
		t.Fatalf("release: %v", err)
	}
	if balance := ledgertest.BalanceOf(t, pool, buyerCard); balance != (ledgertest.Balance{Available: 6 * oneNC, Held: 0}) {
		t.Fatalf("buyer card after release %+v", balance)
	}
	if err := ledgertest.InTransaction(t, pool, func(tx pgx.Tx) error {
		return ledger.Release(context.Background(), tx, buyerHold, 1)
	}); !errors.Is(err, ledger.ErrHoldNotAvailable) {
		t.Fatalf("releasing a finished hold: %v", err)
	}
	ledgertest.RequireIntegrity(t, pool)
}

func TestSpendUsesBucketsInTheFixedOrder(t *testing.T) {
	pool := databasetest.NewPool(t)
	player := ledgertest.CreatePlayer(t, pool)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketCrypto), 5*oneNC)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketEarned), 2*oneNC)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketEarnedPending), oneNC)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketCard), oneNC)

	spend := func(amount int64, referenceID string) error {
		return ledgertest.InTransaction(t, pool, func(tx pgx.Tx) error {
			debitLegs, err := ledger.SpendLegs(context.Background(), tx, player, amount)
			if err != nil {
				return err
			}
			_, err = ledger.Post(context.Background(), tx, ledger.Journal{
				Type:      ledger.JournalShopPurchase,
				Reference: ledger.Reference{Type: "payment", ID: referenceID},
				Legs:      append(debitLegs, ledger.Credit(ledger.System(ledger.SystemTreasury, ledger.NC), amount)),
			})
			return err
		})
	}

	if err := spend(3_500_000, "first"); err != nil {
		t.Fatalf("spend: %v", err)
	}
	expectedAfterSpend := map[ledger.Bucket]int64{
		ledger.BucketCard:          0,
		ledger.BucketEarnedPending: 0,
		ledger.BucketEarned:        500_000,
		ledger.BucketCrypto:        5 * oneNC,
	}
	for bucket, expectedAvailable := range expectedAfterSpend {
		if balance := ledgertest.BalanceOf(t, pool, ledger.PlayerNC(player, bucket)); balance.Available != expectedAvailable {
			t.Fatalf("%s holds %d, want %d", bucket, balance.Available, expectedAvailable)
		}
	}
	if err := spend(6*oneNC, "too-much"); !ledger.IsInsufficientBalance(err) {
		t.Fatalf("overspend: %v", err)
	}
	ledgertest.RequireIntegrity(t, pool)
}

func TestIntegrityCheckDetectsTampering(t *testing.T) {
	pool := databasetest.NewPool(t)
	player := ledgertest.CreatePlayer(t, pool)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketCard), oneNC)
	if _, err := hold(t, pool, ledger.PlayerNC(player, ledger.BucketCard), 100, "tamper"); err != nil {
		t.Fatalf("hold: %v", err)
	}
	ledgertest.RequireIntegrity(t, pool)

	if _, err := pool.Exec(context.Background(), `UPDATE ledger_balances SET available = available + 7, held = held + 1 WHERE available > 0`); err != nil {
		t.Fatalf("tamper: %v", err)
	}
	if _, err := pool.Exec(context.Background(), `UPDATE ledger_entries SET amount = amount + 1 WHERE id = (SELECT min(id) FROM ledger_entries)`); err != nil {
		t.Fatalf("tamper entries: %v", err)
	}
	report, err := ledger.RunIntegrityCheck(context.Background(), pool)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if len(report.UnbalancedJournals) == 0 || len(report.BalancesNotMatchingEntry) == 0 || len(report.HeldNotMatchingHolds) == 0 {
		t.Fatalf("every kind of tampering should be reported: %+v", report)
	}
}
