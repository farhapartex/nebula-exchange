package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

var ErrCoinPackNotFound = errors.New("coin pack not found")

type CoinPackRepository interface {
	ListActive(ctx context.Context, afterSortOrder int, limit int) ([]models.CoinPack, error)
	FindActive(ctx context.Context, coinPackID string) (models.CoinPack, error)
}

type GormCoinPackRepository struct {
	database *gorm.DB
}

func NewCoinPackRepository(database *gorm.DB) *GormCoinPackRepository {
	return &GormCoinPackRepository{database: database}
}

func (repository *GormCoinPackRepository) ListActive(ctx context.Context, afterSortOrder int, limit int) ([]models.CoinPack, error) {
	var coinPacks []models.CoinPack
	err := database.Session(ctx, repository.database).
		Where(map[string]any{"is_active": true}).
		Where("sort_order > ?", afterSortOrder).
		Order("sort_order").
		Limit(limit).
		Find(&coinPacks).Error
	return coinPacks, err
}

func (repository *GormCoinPackRepository) FindActive(ctx context.Context, coinPackID string) (models.CoinPack, error) {
	var coinPack models.CoinPack
	err := database.Session(ctx, repository.database).
		Where(map[string]any{"id": coinPackID, "is_active": true}).
		Take(&coinPack).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.CoinPack{}, ErrCoinPackNotFound
	}
	return coinPack, err
}
