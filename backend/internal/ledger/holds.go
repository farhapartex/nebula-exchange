package ledger

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nebula-exchange/backend/internal/ledger/ledgerstore"
)

func Hold(ctx context.Context, tx pgx.Tx, account AccountKey, amount int64, reference Reference) (uuid.UUID, error) {
	if amount <= 0 {
		return uuid.Nil, ErrInvalidAmount
	}
	queries := ledgerstore.New(tx)
	accountID, err := resolveAccount(ctx, queries, account)
	if err != nil {
		return uuid.Nil, err
	}
	holdID, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("generate hold id: %w", err)
	}
	if _, err := queries.InsertHold(ctx, ledgerstore.InsertHoldParams{
		ID:        holdID,
		AccountID: accountID,
		Amount:    amount,
		RefType:   reference.Type,
		RefID:     reference.ID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrDuplicateHold
		}
		return uuid.Nil, fmt.Errorf("insert hold: %w", err)
	}
	if _, err := queries.MoveAvailableToHeld(ctx, ledgerstore.MoveAvailableToHeldParams{AccountID: accountID, Amount: amount}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, &InsufficientBalanceError{Asset: account.Asset}
		}
		return uuid.Nil, fmt.Errorf("move balance to held: %w", err)
	}
	return holdID, nil
}

func Release(ctx context.Context, tx pgx.Tx, holdID uuid.UUID, amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	queries := ledgerstore.New(tx)
	reducedHold, err := queries.ReduceHold(ctx, ledgerstore.ReduceHoldParams{ID: holdID, Amount: amount, FinalStatus: "RELEASED"})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrHoldNotAvailable
	}
	if err != nil {
		return fmt.Errorf("release hold: %w", err)
	}
	if _, err := queries.MoveHeldToAvailable(ctx, ledgerstore.MoveHeldToAvailableParams{AccountID: reducedHold.AccountID, Amount: amount}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrHoldNotAvailable
		}
		return fmt.Errorf("move held balance back: %w", err)
	}
	return nil
}

func ReleaseRemaining(ctx context.Context, tx pgx.Tx, holdID uuid.UUID) (int64, error) {
	hold, err := ledgerstore.New(tx).GetHold(ctx, holdID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrHoldNotAvailable
	}
	if err != nil {
		return 0, fmt.Errorf("load hold: %w", err)
	}
	if hold.Status != "ACTIVE" {
		return 0, nil
	}
	if err := Release(ctx, tx, holdID, hold.Remaining); err != nil {
		return 0, err
	}
	return hold.Remaining, nil
}
