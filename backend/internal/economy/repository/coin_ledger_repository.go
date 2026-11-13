package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/economy/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type CoinLedgerRepository interface {
	Append(ctx context.Context, entry *models.CoinLedgerEntry) error
	HasWelcomeGift(ctx context.Context, userID uuid.UUID) (bool, error)
	HasPaymentEntry(ctx context.Context, paymentID uuid.UUID, reason models.CoinEntryReason) (bool, error)
}

type GormCoinLedgerRepository struct {
	database *gorm.DB
}

func NewCoinLedgerRepository(database *gorm.DB) *GormCoinLedgerRepository {
	return &GormCoinLedgerRepository{database: database}
}

func (repository *GormCoinLedgerRepository) Append(ctx context.Context, entry *models.CoinLedgerEntry) error {
	return database.Session(ctx, repository.database).Create(entry).Error
}

func (repository *GormCoinLedgerRepository) HasWelcomeGift(ctx context.Context, userID uuid.UUID) (bool, error) {
	return repository.exists(ctx, map[string]any{"user_id": userID, "reason": models.CoinEntryReasonWelcomeGift})
}

func (repository *GormCoinLedgerRepository) HasPaymentEntry(ctx context.Context, paymentID uuid.UUID, reason models.CoinEntryReason) (bool, error) {
	return repository.exists(ctx, map[string]any{"payment_id": paymentID, "reason": reason})
}

func (repository *GormCoinLedgerRepository) exists(ctx context.Context, conditions map[string]any) (bool, error) {
	var matchingEntries int64
	err := database.Session(ctx, repository.database).
		Model(&models.CoinLedgerEntry{}).
		Where(conditions).
		Limit(1).
		Count(&matchingEntries).Error
	return matchingEntries > 0, err
}
