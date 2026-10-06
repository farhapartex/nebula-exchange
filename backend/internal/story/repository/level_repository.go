package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/story/models"
)

type LevelRepository interface {
	FindPublished(ctx context.Context, levelID string) (models.Level, bool, error)
}

type GormLevelRepository struct {
	database *gorm.DB
}

func NewLevelRepository(database *gorm.DB) *GormLevelRepository {
	return &GormLevelRepository{database: database}
}

func (repository *GormLevelRepository) FindPublished(ctx context.Context, levelID string) (models.Level, bool, error) {
	var level models.Level
	err := database.Session(ctx, repository.database).
		Preload("Chapter").
		Where(map[string]any{"id": levelID, "is_published": true}).
		Take(&level).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Level{}, false, nil
	}
	if err != nil {
		return models.Level{}, false, err
	}
	return level, true, nil
}
