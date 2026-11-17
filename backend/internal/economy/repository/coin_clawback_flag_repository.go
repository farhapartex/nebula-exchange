package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/economy/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type CoinClawbackFlagRepository interface {
	Create(ctx context.Context, flag *models.CoinClawbackFlag) error
	ExistsForPayment(ctx context.Context, paymentID uuid.UUID) (bool, error)
}

type GormCoinClawbackFlagRepository struct {
	database *gorm.DB
}

func NewCoinClawbackFlagRepository(database *gorm.DB) *GormCoinClawbackFlagRepository {
	return &GormCoinClawbackFlagRepository{database: database}
}

func (repository *GormCoinClawbackFlagRepository) Create(ctx context.Context, flag *models.CoinClawbackFlag) error {
	return database.Session(ctx, repository.database).Create(flag).Error
}

func (repository *GormCoinClawbackFlagRepository) ExistsForPayment(ctx context.Context, paymentID uuid.UUID) (bool, error) {
	var matchingFlags int64
	err := database.Session(ctx, repository.database).
		Model(&models.CoinClawbackFlag{}).
		Where(map[string]any{"payment_id": paymentID}).
		Limit(1).
		Count(&matchingFlags).Error
	return matchingFlags > 0, err
}
