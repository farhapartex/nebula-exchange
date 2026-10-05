package outbox

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type Repository interface {
	Insert(ctx context.Context, outboxEmail *OutboxEmail) error
	ClaimDue(ctx context.Context, now, claimedUntil time.Time, batchSize int) ([]OutboxEmail, error)
	MarkSent(ctx context.Context, emailID uuid.UUID, now time.Time) error
	RecordFailure(ctx context.Context, failedEmail OutboxEmail, nextAttemptAt time.Time, lastError string, now time.Time) error
}

type GormRepository struct {
	database *gorm.DB
}

func NewGormRepository(database *gorm.DB) *GormRepository {
	return &GormRepository{database: database}
}

func (repository *GormRepository) Insert(ctx context.Context, outboxEmail *OutboxEmail) error {
	return database.Session(ctx, repository.database).Create(outboxEmail).Error
}

func (repository *GormRepository) ClaimDue(ctx context.Context, now, claimedUntil time.Time, batchSize int) ([]OutboxEmail, error) {
	session := database.Session(ctx, repository.database)
	dueEmailIDs := session.Model(&OutboxEmail{}).
		Select("id").
		Where(map[string]any{"status": EmailStatusPending}).
		Where("next_attempt_at <= ?", now).
		Where("claimed_until IS NULL OR claimed_until < ?", now).
		Order("next_attempt_at").
		Limit(batchSize).
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate, Options: clause.LockingOptionsSkipLocked})

	var claimedEmails []OutboxEmail
	err := session.Model(&claimedEmails).
		Clauses(clause.Returning{}).
		Where("id IN (?)", dueEmailIDs).
		Updates(map[string]any{"claimed_until": claimedUntil, "updated_at": now}).Error
	return claimedEmails, err
}

func (repository *GormRepository) MarkSent(ctx context.Context, emailID uuid.UUID, now time.Time) error {
	return database.Session(ctx, repository.database).
		Model(&OutboxEmail{}).
		Where(map[string]any{"id": emailID}).
		Updates(map[string]any{
			"status":        EmailStatusSent,
			"attempts":      gorm.Expr("attempts + 1"),
			"sent_at":       now,
			"claimed_until": nil,
			"last_error":    nil,
			"updated_at":    now,
		}).Error
}

func (repository *GormRepository) RecordFailure(ctx context.Context, failedEmail OutboxEmail, nextAttemptAt time.Time, lastError string, now time.Time) error {
	failedAttemptCount := failedEmail.Attempts + 1
	nextStatus := EmailStatusPending
	if failedAttemptCount >= failedEmail.MaxAttempts {
		nextStatus = EmailStatusFailed
	}
	return database.Session(ctx, repository.database).
		Model(&OutboxEmail{}).
		Where(map[string]any{"id": failedEmail.ID}).
		Updates(map[string]any{
			"status":          nextStatus,
			"attempts":        failedAttemptCount,
			"next_attempt_at": nextAttemptAt,
			"last_error":      lastError,
			"claimed_until":   nil,
			"updated_at":      now,
		}).Error
}
