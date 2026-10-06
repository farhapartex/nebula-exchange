package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/wallet/models"
)

type ChallengeClaim struct {
	UserID  uuid.UUID
	Address string
	ChainID int64
	Nonce   string
	UsedAt  time.Time
}

type WalletChallengeRepository interface {
	Create(ctx context.Context, challenge *models.WalletChallenge) error
	Consume(ctx context.Context, claim ChallengeClaim) (bool, error)
}

type GormWalletChallengeRepository struct {
	database *gorm.DB
}

func NewWalletChallengeRepository(database *gorm.DB) *GormWalletChallengeRepository {
	return &GormWalletChallengeRepository{database: database}
}

func (repository *GormWalletChallengeRepository) Create(ctx context.Context, challenge *models.WalletChallenge) error {
	return database.Session(ctx, repository.database).Create(challenge).Error
}

func (repository *GormWalletChallengeRepository) Consume(ctx context.Context, claim ChallengeClaim) (bool, error) {
	result := database.Session(ctx, repository.database).
		Model(&models.WalletChallenge{}).
		Where(map[string]any{"nonce": claim.Nonce, "user_id": claim.UserID, "address": claim.Address, "chain_id": claim.ChainID}).
		Where("used_at IS NULL AND expires_at > ?", claim.UsedAt).
		Update("used_at", claim.UsedAt)
	return result.RowsAffected == 1, result.Error
}
