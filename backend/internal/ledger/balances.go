package ledger

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"nebula-exchange/backend/internal/ledger/ledgerstore"
)

func AccountBalance(ctx context.Context, db ledgerstore.DBTX, account AccountKey) (available int64, held int64, err error) {
	if err := account.validate(); err != nil {
		return 0, 0, err
	}
	queries := ledgerstore.New(db)
	accountID, err := queries.FindAccount(ctx, account.storeParameters())
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, 0, nil
	}
	if err != nil {
		return 0, 0, fmt.Errorf("find ledger account: %w", err)
	}
	balance, err := queries.GetBalance(ctx, accountID)
	if err != nil {
		return 0, 0, fmt.Errorf("read balance: %w", err)
	}
	return balance.Available, balance.Held, nil
}
