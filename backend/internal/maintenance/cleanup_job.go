package maintenance

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/maintenance/maintenancestore"
)

const (
	neverActivatedAccountLifetime = 7 * 24 * time.Hour
	expiredTokenRetention         = 7 * 24 * time.Hour
	revokedSessionRetention       = 30 * 24 * time.Hour
	idempotencyKeyRetention       = 24 * time.Hour
	sentEmailRetention            = 30 * 24 * time.Hour
)

type CleanupJob struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
	now    func() time.Time
}

func NewCleanupJob(pool *pgxpool.Pool, logger *slog.Logger, now func() time.Time) *CleanupJob {
	return &CleanupJob{pool: pool, logger: logger, now: now}
}

func (job *CleanupJob) Name() string            { return "cleanup" }
func (job *CleanupJob) Interval() time.Duration { return 24 * time.Hour }

type cleanupStep struct {
	name    string
	execute func(ctx context.Context, queries *maintenancestore.Queries, now time.Time) (int64, error)
}

var cleanupSteps = []cleanupStep{
	{"never_activated_users", func(ctx context.Context, queries *maintenancestore.Queries, now time.Time) (int64, error) {
		return queries.DeleteNeverActivatedUsers(ctx, now.Add(-neverActivatedAccountLifetime))
	}},
	{"expired_activation_tokens", func(ctx context.Context, queries *maintenancestore.Queries, now time.Time) (int64, error) {
		return queries.DeleteExpiredActivationTokens(ctx, now.Add(-expiredTokenRetention))
	}},
	{"expired_password_reset_tokens", func(ctx context.Context, queries *maintenancestore.Queries, now time.Time) (int64, error) {
		return queries.DeleteExpiredPasswordResetTokens(ctx, now.Add(-expiredTokenRetention))
	}},
	{"stale_refresh_tokens", func(ctx context.Context, queries *maintenancestore.Queries, now time.Time) (int64, error) {
		return queries.DeleteStaleRefreshTokens(ctx, now.Add(-revokedSessionRetention))
	}},
	{"old_idempotency_keys", func(ctx context.Context, queries *maintenancestore.Queries, now time.Time) (int64, error) {
		return queries.DeleteOldIdempotencyKeys(ctx, now.Add(-idempotencyKeyRetention))
	}},
	{"old_sent_emails", func(ctx context.Context, queries *maintenancestore.Queries, now time.Time) (int64, error) {
		return queries.DeleteOldSentEmails(ctx, now.Add(-sentEmailRetention))
	}},
}

func (job *CleanupJob) Run(ctx context.Context) error {
	queries := maintenancestore.New(job.pool)
	now := job.now().UTC()
	for _, step := range cleanupSteps {
		deletedRows, err := step.execute(ctx, queries, now)
		if err != nil {
			return err
		}
		if deletedRows > 0 {
			job.logger.InfoContext(ctx, "cleanup removed rows", slog.String("step", step.name), slog.Int64("rows", deletedRows))
		}
	}
	return nil
}
