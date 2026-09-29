package balances

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/ledger/ledgerstore"
	"nebula-exchange/backend/internal/platform/money"
)

type BucketBalance struct {
	Bucket    ledger.Bucket `json:"bucket"`
	Available money.Micro   `json:"available"`
	Held      money.Micro   `json:"held"`
}

type Summary struct {
	Total        money.Micro     `json:"total"`
	Available    money.Micro     `json:"available"`
	Held         money.Micro     `json:"held"`
	Withdrawable money.Micro     `json:"withdrawable"`
	Buckets      []BucketBalance `json:"buckets"`
}

var withdrawableBuckets = map[ledger.Bucket]bool{ledger.BucketEarned: true, ledger.BucketCrypto: true}

type Reader struct {
	pool *pgxpool.Pool
}

func NewReader(pool *pgxpool.Pool) *Reader {
	return &Reader{pool: pool}
}

func (reader *Reader) Summarize(ctx context.Context, userID uuid.UUID) (Summary, error) {
	balanceRows, err := ledgerstore.New(reader.pool).ListPlayerNCBalances(ctx, &userID)
	if err != nil {
		return Summary{}, fmt.Errorf("load NC balances: %w", err)
	}
	balancesByBucket := map[ledger.Bucket]ledgerstore.ListPlayerNCBalancesRow{}
	for _, balanceRow := range balanceRows {
		if balanceRow.Bucket != nil {
			balancesByBucket[ledger.Bucket(*balanceRow.Bucket)] = balanceRow
		}
	}

	summary := Summary{Buckets: make([]BucketBalance, 0, len(ledger.SpendingOrder))}
	for _, bucket := range ledger.SpendingOrder {
		balanceRow := balancesByBucket[bucket]
		summary.Buckets = append(summary.Buckets, BucketBalance{
			Bucket:    bucket,
			Available: money.Micro(balanceRow.Available),
			Held:      money.Micro(balanceRow.Held),
		})
		summary.Available += money.Micro(balanceRow.Available)
		summary.Held += money.Micro(balanceRow.Held)
		if withdrawableBuckets[bucket] {
			summary.Withdrawable += money.Micro(max(balanceRow.Available, 0))
		}
	}
	summary.Total = summary.Available + summary.Held
	return summary, nil
}
