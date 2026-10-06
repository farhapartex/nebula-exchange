package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
)

type ChapterUnlockRepository interface {
	ListUnlockedChapterIDs(ctx context.Context, userID uuid.UUID) ([]string, error)
}

type GormChapterUnlockRepository struct {
	database *gorm.DB
}

func NewChapterUnlockRepository(database *gorm.DB) *GormChapterUnlockRepository {
	return &GormChapterUnlockRepository{database: database}
}

func (repository *GormChapterUnlockRepository) ListUnlockedChapterIDs(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var chapterIDs []string
	err := database.Session(ctx, repository.database).
		Model(&models.ChapterUnlock{}).
		Where(map[string]any{"user_id": userID}).
		Pluck("chapter_id", &chapterIDs).Error
	return chapterIDs, err
}
