package purpose

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/platform/money"
	"nebula-exchange/backend/internal/users"
)

const EntryFee = 5 * money.MicroPerNC

const (
	scoutItemID      = 301
	drillTierOneID   = 201
	fuelCellItemID   = 401
	starterFuelCells = 10
	starterCardBonus = 2 * money.MicroPerNC
)

type EntryFeeHandler struct {
	users *users.Repository
	now   func() time.Time
}

func NewEntryFeeHandler(userRepository *users.Repository, now func() time.Time) *EntryFeeHandler {
	return &EntryFeeHandler{users: userRepository, now: now}
}

func (handler *EntryFeeHandler) Apply(ctx context.Context, tx pgx.Tx, payment Payment) error {
	isActivated, err := handler.users.ChangeStatus(ctx, tx, payment.UserID, users.StatusPendingPayment, users.StatusActive, handler.now())
	if err != nil {
		return fmt.Errorf("activate account: %w", err)
	}
	if !isActivated {
		return &FailureError{Code: "ACCOUNT_NOT_PENDING_PAYMENT"}
	}

	feeLegs, err := ledger.SpendLegs(ctx, tx, payment.UserID, int64(EntryFee))
	if err != nil {
		return err
	}
	if _, err := ledger.Post(ctx, tx, ledger.Journal{
		Type:      ledger.JournalEntryFee,
		Reference: paymentReference(payment),
		Legs:      append(feeLegs, ledger.Credit(ledger.System(ledger.SystemTreasury, ledger.NC), int64(EntryFee))),
	}); err != nil {
		return err
	}
	return GrantStarterPack(ctx, tx, payment)
}

func GrantStarterPack(ctx context.Context, tx pgx.Tx, payment Payment) error {
	legs := []ledger.Leg{
		ledger.Debit(ledger.System(ledger.SystemMint, ledger.NC), int64(starterCardBonus)),
		ledger.Credit(ledger.PlayerNC(payment.UserID, ledger.BucketCard), int64(starterCardBonus)),
	}
	for _, starterItem := range []ledger.Leg{
		{Account: ledger.PlayerItem(payment.UserID, scoutItemID), Amount: 1},
		{Account: ledger.PlayerItem(payment.UserID, drillTierOneID), Amount: 1},
		{Account: ledger.PlayerItem(payment.UserID, fuelCellItemID), Amount: starterFuelCells},
	} {
		legs = append(legs, ledger.Debit(ledger.System(ledger.SystemMint, starterItem.Account.Asset), starterItem.Amount), starterItem)
	}
	_, err := ledger.Post(ctx, tx, ledger.Journal{
		Type:      ledger.JournalStarterPack,
		Reference: ledger.Reference{Type: "user", ID: payment.UserID.String()},
		Metadata:  map[string]any{"payment_id": payment.ID.String()},
		Legs:      legs,
	})
	return err
}
