package ledger_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5"

	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/ledger/ledgertest"
	"nebula-exchange/backend/internal/platform/database/databasetest"
)

func runConcurrently(attempts int, attempt func(attemptIndex int) error) (succeeded int32, unexpected []error) {
	var waitGroup sync.WaitGroup
	var successCount atomic.Int32
	var unexpectedMutex sync.Mutex
	for attemptIndex := 0; attemptIndex < attempts; attemptIndex++ {
		waitGroup.Add(1)
		go func(index int) {
			defer waitGroup.Done()
			err := attempt(index)
			switch {
			case err == nil:
				successCount.Add(1)
			case ledger.IsInsufficientBalance(err):
			default:
				unexpectedMutex.Lock()
				unexpected = append(unexpected, err)
				unexpectedMutex.Unlock()
			}
		}(attemptIndex)
	}
	waitGroup.Wait()
	return successCount.Load(), unexpected
}

func TestParallelSpendsNeverOverspend(t *testing.T) {
	pool := databasetest.NewPool(t)
	player := ledgertest.CreatePlayer(t, pool)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketCard), 6*oneNC)
	ledgertest.Fund(t, pool, ledger.PlayerNC(player, ledger.BucketEarned), 4*oneNC)

	succeeded, unexpected := runConcurrently(40, func(attemptIndex int) error {
		return ledgertest.InTransaction(t, pool, func(tx pgx.Tx) error {
			debitLegs, err := ledger.SpendLegs(context.Background(), tx, player, oneNC)
			if err != nil {
				return err
			}
			_, err = ledger.Post(context.Background(), tx, ledger.Journal{
				Type:      ledger.JournalShopPurchase,
				Reference: ledger.Reference{Type: "payment", ID: fmt.Sprintf("parallel-%d", attemptIndex)},
				Legs:      append(debitLegs, ledger.Credit(ledger.System(ledger.SystemTreasury, ledger.NC), oneNC)),
			})
			return err
		})
	})
	if len(unexpected) > 0 {
		t.Fatalf("unexpected errors: %v", unexpected)
	}
	if succeeded != 10 {
		t.Fatalf("%d spends succeeded, want exactly 10", succeeded)
	}
	for _, bucket := range []ledger.Bucket{ledger.BucketCard, ledger.BucketEarned} {
		if balance := ledgertest.BalanceOf(t, pool, ledger.PlayerNC(player, bucket)); balance.Available != 0 {
			t.Fatalf("%s left with %d", bucket, balance.Available)
		}
	}
	ledgertest.RequireIntegrity(t, pool)
}

func TestParallelHoldsNeverExceedAvailable(t *testing.T) {
	pool := databasetest.NewPool(t)
	player := ledgertest.CreatePlayer(t, pool)
	playerFuel := ledger.PlayerItem(player, fuelCellItemID)
	ledgertest.Fund(t, pool, playerFuel, 10)

	succeeded, unexpected := runConcurrently(30, func(attemptIndex int) error {
		_, err := hold(t, pool, playerFuel, 1, fmt.Sprintf("order-%d", attemptIndex))
		return err
	})
	if len(unexpected) > 0 {
		t.Fatalf("unexpected errors: %v", unexpected)
	}
	if succeeded != 10 {
		t.Fatalf("%d holds succeeded, want 10", succeeded)
	}
	if balance := ledgertest.BalanceOf(t, pool, playerFuel); balance != (ledgertest.Balance{Available: 0, Held: 10}) {
		t.Fatalf("fuel after holds %+v", balance)
	}
	ledgertest.RequireIntegrity(t, pool)
}

func TestOpposingTransfersDoNotDeadlock(t *testing.T) {
	pool := databasetest.NewPool(t)
	firstPlayer := ledgertest.CreatePlayer(t, pool)
	secondPlayer := ledgertest.CreatePlayer(t, pool)
	firstEarned := ledger.PlayerNC(firstPlayer, ledger.BucketEarned)
	secondEarned := ledger.PlayerNC(secondPlayer, ledger.BucketEarned)
	ledgertest.Fund(t, pool, firstEarned, 100*oneNC)
	ledgertest.Fund(t, pool, secondEarned, 100*oneNC)

	succeeded, unexpected := runConcurrently(60, func(attemptIndex int) error {
		source, destination := firstEarned, secondEarned
		if attemptIndex%2 == 1 {
			source, destination = secondEarned, firstEarned
		}
		return post(t, pool, ledger.Journal{
			Type:      ledger.JournalTradeFill,
			Reference: ledger.Reference{Type: "trade", ID: fmt.Sprintf("swap-%d", attemptIndex)},
			Legs: []ledger.Leg{
				ledger.Debit(source, oneNC),
				ledger.Credit(destination, oneNC),
				ledger.Credit(ledger.System(ledger.SystemFees, ledger.NC), 1),
				ledger.Debit(source, 1),
			},
		})
	})
	if len(unexpected) > 0 {
		t.Fatalf("transfers failed (a deadlock shows up here): %v", unexpected)
	}
	if succeeded != 60 {
		t.Fatalf("%d transfers succeeded, want 60", succeeded)
	}
	total := ledgertest.BalanceOf(t, pool, firstEarned).Available + ledgertest.BalanceOf(t, pool, secondEarned).Available
	if total != 200*oneNC-60 {
		t.Fatalf("players hold %d together", total)
	}
	ledgertest.RequireIntegrity(t, pool)
}
