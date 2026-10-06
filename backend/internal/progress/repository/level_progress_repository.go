package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
)

type LevelProgressRepository interface {
	IsCompleted(ctx context.Context, userID uuid.UUID, levelID string) (bool, error)
	RecordStart(ctx context.Context, userID uuid.UUID, levelID string, startedAt time.Time) error
	ListForUser(ctx context.Context, userID uuid.UUID) ([]models.LevelProgress, error)
	RecordWin(ctx context.Context, userID uuid.UUID, levelID string, stars int, wonAt time.Time) error
}

type GormLevelProgressRepository struct {
	database *gorm.DB
}

func NewLevelProgressRepository(database *gorm.DB) *GormLevelProgressRepository {
	return &GormLevelProgressRepository{database: database}
}

func (repository *GormLevelProgressRepository) IsCompleted(ctx context.Context, userID uuid.UUID, levelID string) (bool, error) {
	var completedCount int64
	err := database.Session(ctx, repository.database).
		Model(&models.LevelProgress{}).
		Where(map[string]any{"user_id": userID, "level_id": levelID, "status": models.ProgressStatusCompleted}).
		Count(&completedCount).Error
	return completedCount > 0, err
}

func (repository *GormLevelProgressRepository) RecordStart(ctx context.Context, userID uuid.UUID, levelID string, startedAt time.Time) error {
	return database.Session(ctx, repository.database).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "level_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"attempts":   gorm.Expr("level_progress.attempts + 1"),
				"updated_at": startedAt,
			}),
		}).
		Create(&models.LevelProgress{
			UserID:         userID,
			LevelID:        levelID,
			Status:         models.ProgressStatusStarted,
			Attempts:       1,
			FirstStartedAt: startedAt,
			CreatedAt:      startedAt,
			UpdatedAt:      startedAt,
		}).Error
}

func (repository *GormLevelProgressRepository) ListForUser(ctx context.Context, userID uuid.UUID) ([]models.LevelProgress, error) {
	var levelProgress []models.LevelProgress
	err := database.Session(ctx, repository.database).
		Where(map[string]any{"user_id": userID}).
		Find(&levelProgress).Error
	return levelProgress, err
}

func (repository *GormLevelProgressRepository) RecordWin(ctx context.Context, userID uuid.UUID, levelID string, stars int, wonAt time.Time) error {
	return database.Session(ctx, repository.database).
		Model(&models.LevelProgress{}).
		Where(map[string]any{"user_id": userID, "level_id": levelID}).
		Updates(map[string]any{
			"status":             models.ProgressStatusCompleted,
			"first_completed_at": gorm.Expr("COALESCE(first_completed_at, ?)", wonAt),
			"best_stars":         gorm.Expr("GREATEST(COALESCE(best_stars, 0), ?)", stars),
			"updated_at":         wonAt,
		}).Error
}
