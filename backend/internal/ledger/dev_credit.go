package ledger

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func DevCredit(ctx context.Context, tx pgx.Tx, account AccountKey, amount int64) (uuid.UUID, error) {
	if amount <= 0 {
		return uuid.Nil, ErrInvalidAmount
	}
	creditID, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}
	return Post(ctx, tx, Journal{
		Type:      JournalDevCredit,
		Reference: Reference{Type: "dev_credit", ID: creditID.String()},
		Legs: []Leg{
			Debit(System(SystemMint, account.Asset), amount),
			Credit(account, amount),
		},
	})
}
