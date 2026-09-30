package payments

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/payments/paymentsstore"
)

type ExpiryJob struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
	now    func() time.Time
}

func NewExpiryJob(pool *pgxpool.Pool, logger *slog.Logger, now func() time.Time) *ExpiryJob {
	return &ExpiryJob{pool: pool, logger: logger, now: now}
}

func (job *ExpiryJob) Name() string            { return "payment_expiry" }
func (job *ExpiryJob) Interval() time.Duration { return time.Minute }

func (job *ExpiryJob) Run(ctx context.Context) error {
	expiredPaymentIDs, err := paymentsstore.New(job.pool).ExpirePendingPayments(ctx, job.now())
	if err != nil {
		return fmt.Errorf("expire pending payments: %w", err)
	}
	if len(expiredPaymentIDs) > 0 {
		job.logger.Info("expired pending payments", slog.Int("count", len(expiredPaymentIDs)))
	}
	return nil
}
