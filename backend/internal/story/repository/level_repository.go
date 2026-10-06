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
	ListPublishedStoryLevels(ctx context.Context) ([]models.Level, error)
	FindWithFightContent(ctx context.Context, levelID string) (models.Level, bool, error)
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

func (repository *GormLevelRepository) ListPublishedStoryLevels(ctx context.Context) ([]models.Level, error) {
	var levels []models.Level
	err := database.Session(ctx, repository.database).
		Joins("Chapter").
		Where(map[string]any{"levels.is_published": true, "Chapter.is_published": true}).
		Where("levels.kind <> ?", models.LevelKindTraining).
		Order("\"Chapter\".number").
		Order("levels.number").
		Find(&levels).Error
	return levels, err
}

func (repository *GormLevelRepository) FindWithFightContent(ctx context.Context, levelID string) (models.Level, bool, error) {
	var level models.Level
	err := database.Session(ctx, repository.database).
		Preload("Arena").
		Preload("Enemies", func(query *gorm.DB) *gorm.DB { return query.Order("wave") }).
		Preload("Enemies.Enemy").
		Where(map[string]any{"id": levelID}).
		Take(&level).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Level{}, false, nil
	}
	if err != nil {
		return models.Level{}, false, err
	}
	return level, true, nil
}
