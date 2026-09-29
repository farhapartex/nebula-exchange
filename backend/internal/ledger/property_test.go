package ledger_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/ledger/ledgertest"
	"nebula-exchange/backend/internal/platform/database/databasetest"
)

type modelHold struct {
	id        uuid.UUID
	account   ledger.AccountKey
	remaining int64
}

type ledgerModel struct {
	balances map[ledger.AccountKey]ledgertest.Balance
	holds    []*modelHold
}

func (model *ledgerModel) activeHolds() []*modelHold {
	active := []*modelHold{}
	for _, candidate := range model.holds {
		if candidate.remaining > 0 {
			active = append(active, candidate)
		}
	}
	return active
}

func (model *ledgerModel) adjust(account ledger.AccountKey, availableChange, heldChange int64) {
	balance := model.balances[account]
	balance.Available += availableChange
	balance.Held += heldChange
	model.balances[account] = balance
}

func TestRandomOperationsMatchAModelAndKeepIntegrity(t *testing.T) {
	pool := databasetest.NewPool(t)
	randomSource := rand.New(rand.NewPCG(20260930, 42))
	players := []uuid.UUID{ledgertest.CreatePlayer(t, pool), ledgertest.CreatePlayer(t, pool), ledgertest.CreatePlayer(t, pool)}

	var accounts []ledger.AccountKey
	for _, player := range players {
		for _, bucket := range ledger.SpendingOrder {
			accounts = append(accounts, ledger.PlayerNC(player, bucket))
		}
		accounts = append(accounts, ledger.PlayerItem(player, fuelCellItemID))
	}
	model := &ledgerModel{balances: map[ledger.AccountKey]ledgertest.Balance{}}
	pickAccount := func() ledger.AccountKey { return accounts[randomSource.IntN(len(accounts))] }

	for operationIndex := 0; operationIndex < 300; operationIndex++ {
		referenceID := fmt.Sprintf("op-%d", operationIndex)
		switch randomSource.IntN(5) {
		case 0:
			account := pickAccount()
			amount := randomSource.Int64N(50) + 1
			ledgertest.Fund(t, pool, account, amount)
			model.adjust(account, amount, 0)

		case 1:
			source := pickAccount()
			destination := pickAccount()
			for destination.Asset != source.Asset || destination == source {
				destination = pickAccount()
			}
			amount := randomSource.Int64N(60) + 1
			err := post(t, pool, ledger.Journal{
				Type:      ledger.JournalTradeFill,
				Reference: ledger.Reference{Type: "random", ID: referenceID},
				Legs:      []ledger.Leg{ledger.Debit(source, amount), ledger.Credit(destination, amount)},
			})
			assertOutcome(t, operationIndex, err, model.balances[source].Available < amount)
			if err == nil {
				model.adjust(source, -amount, 0)
				model.adjust(destination, amount, 0)
			}

		case 2:
			account := pickAccount()
			amount := randomSource.Int64N(40) + 1
			holdID, err := hold(t, pool, account, amount, referenceID)
			assertOutcome(t, operationIndex, err, model.balances[account].Available < amount)
			if err == nil {
				model.adjust(account, -amount, amount)
				model.holds = append(model.holds, &modelHold{id: holdID, account: account, remaining: amount})
			}

		case 3:
			activeHolds := model.activeHolds()
			if len(activeHolds) == 0 {
				continue
			}
			releasedHold := activeHolds[randomSource.IntN(len(activeHolds))]
			amount := randomSource.Int64N(releasedHold.remaining) + 1
			err := ledgertest.InTransaction(t, pool, func(tx pgx.Tx) error {
				return ledger.Release(context.Background(), tx, releasedHold.id, amount)
			})
			assertOutcome(t, operationIndex, err, false)
			releasedHold.remaining -= amount
			model.adjust(releasedHold.account, amount, -amount)

		case 4:
			activeHolds := model.activeHolds()
			if len(activeHolds) == 0 {
				continue
			}
			capturedHold := activeHolds[randomSource.IntN(len(activeHolds))]
			destination := pickAccount()
			for destination.Asset != capturedHold.account.Asset {
				destination = pickAccount()
			}
			amount := randomSource.Int64N(capturedHold.remaining+5) + 1
			err := post(t, pool, ledger.Journal{
				Type:      ledger.JournalTradeFill,
				Reference: ledger.Reference{Type: "random", ID: referenceID},
				Legs:      []ledger.Leg{ledger.Credit(destination, amount)},
			}, ledger.HoldCapture{HoldID: capturedHold.id, Amount: amount})
			assertOutcome(t, operationIndex, err, amount > capturedHold.remaining)
			if err == nil {
				capturedHold.remaining -= amount
				model.adjust(capturedHold.account, 0, -amount)
				model.adjust(destination, amount, 0)
			}
		}
	}

	for account, expected := range model.balances {
		if actual := ledgertest.BalanceOf(t, pool, account); actual != expected {
			t.Fatalf("account %+v is %+v, model says %+v", account, actual, expected)
		}
	}
	ledgertest.RequireIntegrity(t, pool)
}

func assertOutcome(t *testing.T, operationIndex int, err error, expectsFailure bool) {
	t.Helper()
	if expectsFailure && err == nil {
		t.Fatalf("operation %d should have failed", operationIndex)
	}
	if !expectsFailure && err != nil {
		t.Fatalf("operation %d failed unexpectedly: %v", operationIndex, err)
	}
}
