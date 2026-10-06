package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/billing/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type PlanRepository interface {
	ListActive(ctx context.Context, afterSortOrder int, limit int) ([]models.Plan, error)
}

type GormPlanRepository struct {
	database *gorm.DB
}

func NewPlanRepository(database *gorm.DB) *GormPlanRepository {
	return &GormPlanRepository{database: database}
}

func (repository *GormPlanRepository) ListActive(ctx context.Context, afterSortOrder int, limit int) ([]models.Plan, error) {
	var plans []models.Plan
	err := database.Session(ctx, repository.database).
		Where(map[string]any{"is_active": true}).
		Where("sort_order > ?", afterSortOrder).
		Order("sort_order").
		Limit(limit).
		Find(&plans).Error
	return plans, err
}
