package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type ActivationTokenRepository interface {
	Create(ctx context.Context, activationToken *models.ActivationToken) error
	FindByHash(ctx context.Context, tokenHash []byte) (models.ActivationToken, bool, error)
	Consume(ctx context.Context, tokenHash []byte, now time.Time) (models.ActivationToken, bool, error)
}

type GormActivationTokenRepository struct {
	database *gorm.DB
}

func NewActivationTokenRepository(database *gorm.DB) *GormActivationTokenRepository {
	return &GormActivationTokenRepository{database: database}
}

func (repository *GormActivationTokenRepository) Create(ctx context.Context, activationToken *models.ActivationToken) error {
	return database.Session(ctx, repository.database).Create(activationToken).Error
}

func (repository *GormActivationTokenRepository) FindByHash(ctx context.Context, tokenHash []byte) (models.ActivationToken, bool, error) {
	var storedToken models.ActivationToken
	err := database.Session(ctx, repository.database).Where("token_hash = ?", tokenHash).Take(&storedToken).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.ActivationToken{}, false, nil
	}
	if err != nil {
		return models.ActivationToken{}, false, err
	}
	return storedToken, true, nil
}

func (repository *GormActivationTokenRepository) Consume(ctx context.Context, tokenHash []byte, now time.Time) (models.ActivationToken, bool, error) {
	var consumedTokens []models.ActivationToken
	err := database.Session(ctx, repository.database).
		Model(&consumedTokens).
		Clauses(clause.Returning{}).
		Where("token_hash = ?", tokenHash).
		Where("used_at IS NULL AND expires_at > ?", now).
		Update("used_at", now).Error
	if err != nil || len(consumedTokens) == 0 {
		return models.ActivationToken{}, false, err
	}
	return consumedTokens[0], true, nil
}
