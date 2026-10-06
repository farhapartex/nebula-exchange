package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type PlanRepository interface {
	ListActive(ctx context.Context, afterSortOrder int, limit int) ([]models.Plan, error)
	FindActive(ctx context.Context, planID string) (models.Plan, error)
}

var ErrPlanNotFound = errors.New("plan not found")

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

func (repository *GormPlanRepository) FindActive(ctx context.Context, planID string) (models.Plan, error) {
	var plan models.Plan
	err := database.Session(ctx, repository.database).
		Where(map[string]any{"id": planID, "is_active": true}).
		Take(&plan).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Plan{}, ErrPlanNotFound
	}
	return plan, err
}
