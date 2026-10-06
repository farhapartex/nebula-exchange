package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type StripeWebhookEventRepository interface {
	RecordIfNew(ctx context.Context, event *models.StripeWebhookEvent) (bool, error)
}

type GormStripeWebhookEventRepository struct {
	database *gorm.DB
}

func NewStripeWebhookEventRepository(database *gorm.DB) *GormStripeWebhookEventRepository {
	return &GormStripeWebhookEventRepository{database: database}
}

func (repository *GormStripeWebhookEventRepository) RecordIfNew(ctx context.Context, event *models.StripeWebhookEvent) (bool, error) {
	result := database.Session(ctx, repository.database).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(event)
	return result.RowsAffected == 1, result.Error
}
