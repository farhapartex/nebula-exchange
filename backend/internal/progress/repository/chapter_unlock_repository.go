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

type ChapterUnlockRepository interface {
	ListUnlockedChapterIDs(ctx context.Context, userID uuid.UUID) ([]string, error)
	UnlockForPayment(ctx context.Context, userID uuid.UUID, paymentID uuid.UUID, chapterIDs []string, unlockedAt time.Time) error
	LockForPayment(ctx context.Context, paymentID uuid.UUID) error
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

func (repository *GormChapterUnlockRepository) UnlockForPayment(ctx context.Context, userID uuid.UUID, paymentID uuid.UUID, chapterIDs []string, unlockedAt time.Time) error {
	if len(chapterIDs) == 0 {
		return nil
	}
	chapterUnlocks := make([]models.ChapterUnlock, 0, len(chapterIDs))
	for _, chapterID := range chapterIDs {
		chapterUnlocks = append(chapterUnlocks, models.ChapterUnlock{
			UserID:     userID,
			ChapterID:  chapterID,
			Source:     models.ChapterUnlockSourcePurchase,
			PaymentID:  &paymentID,
			UnlockedAt: unlockedAt,
		})
	}
	return database.Session(ctx, repository.database).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&chapterUnlocks).Error
}

func (repository *GormChapterUnlockRepository) LockForPayment(ctx context.Context, paymentID uuid.UUID) error {
	return database.Session(ctx, repository.database).
		Where(map[string]any{"payment_id": paymentID}).
		Delete(&models.ChapterUnlock{}).Error
}
