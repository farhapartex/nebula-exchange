package ledger

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nebula-exchange/backend/internal/ledger/ledgerstore"
)

func SpendLegs(ctx context.Context, tx pgx.Tx, userID uuid.UUID, amount int64) ([]Leg, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	balanceRows, err := ledgerstore.New(tx).ListPlayerNCBalances(ctx, &userID)
	if err != nil {
		return nil, fmt.Errorf("load NC balances: %w", err)
	}
	availableByBucket := map[Bucket]int64{}
	for _, balanceRow := range balanceRows {
		if balanceRow.Bucket != nil {
			availableByBucket[Bucket(*balanceRow.Bucket)] = balanceRow.Available
		}
	}

	remainingToSpend := amount
	debitLegs := make([]Leg, 0, len(SpendingOrder))
	for _, bucket := range SpendingOrder {
		if remainingToSpend == 0 {
			break
		}
		spendFromBucket := min(max(availableByBucket[bucket], 0), remainingToSpend)
		if spendFromBucket == 0 {
			continue
		}
		debitLegs = append(debitLegs, Debit(PlayerNC(userID, bucket), spendFromBucket))
		remainingToSpend -= spendFromBucket
	}
	if remainingToSpend > 0 {
		return nil, &InsufficientBalanceError{Asset: NC}
	}
	return debitLegs, nil
}
