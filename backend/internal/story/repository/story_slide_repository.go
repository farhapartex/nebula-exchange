package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/story/models"
)

type StorySlideRepository interface {
	ListByLevel(ctx context.Context, levelID string, afterPosition int, limit int) ([]models.StorySlide, error)
}

type GormStorySlideRepository struct {
	database *gorm.DB
}

func NewStorySlideRepository(database *gorm.DB) *GormStorySlideRepository {
	return &GormStorySlideRepository{database: database}
}

func (repository *GormStorySlideRepository) ListByLevel(ctx context.Context, levelID string, afterPosition int, limit int) ([]models.StorySlide, error) {
	var slides []models.StorySlide
	err := database.Session(ctx, repository.database).
		Where(map[string]any{"level_id": levelID}).
		Where("position > ?", afterPosition).
		Order("position").
		Limit(limit).
		Find(&slides).Error
	return slides, err
}
