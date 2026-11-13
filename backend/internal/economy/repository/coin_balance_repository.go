package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/farhapartex/nebula-exchange/backend/internal/economy/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type CoinBalanceRepository interface {
	CoinsOf(ctx context.Context, userID uuid.UUID) (int64, error)
	LockOrCreate(ctx context.Context, userID uuid.UUID, createdAt time.Time) (models.CoinBalance, error)
	SetCoins(ctx context.Context, userID uuid.UUID, coins int64, updatedAt time.Time) error
}

type GormCoinBalanceRepository struct {
	database *gorm.DB
}

func NewCoinBalanceRepository(database *gorm.DB) *GormCoinBalanceRepository {
	return &GormCoinBalanceRepository{database: database}
}

func (repository *GormCoinBalanceRepository) CoinsOf(ctx context.Context, userID uuid.UUID) (int64, error) {
	var balance models.CoinBalance
	err := database.Session(ctx, repository.database).Where(map[string]any{"user_id": userID}).Take(&balance).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return balance.Coins, err
}

func (repository *GormCoinBalanceRepository) LockOrCreate(ctx context.Context, userID uuid.UUID, createdAt time.Time) (models.CoinBalance, error) {
	session := database.Session(ctx, repository.database)
	err := session.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&models.CoinBalance{UserID: userID, Coins: 0, UpdatedAt: createdAt}).Error
	if err != nil {
		return models.CoinBalance{}, err
	}
	var balance models.CoinBalance
	err = database.Session(ctx, repository.database).
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Where(map[string]any{"user_id": userID}).
		Take(&balance).Error
	return balance, err
}

func (repository *GormCoinBalanceRepository) SetCoins(ctx context.Context, userID uuid.UUID, coins int64, updatedAt time.Time) error {
	return database.Session(ctx, repository.database).
		Model(&models.CoinBalance{}).
		Where(map[string]any{"user_id": userID}).
		Updates(map[string]any{"coins": coins, "updated_at": updatedAt}).Error
}
