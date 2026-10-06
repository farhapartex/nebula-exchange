package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/progress/repository"
)

type ChapterUnlockService interface {
	UnlockPurchasedChapters(ctx context.Context, userID uuid.UUID, paymentID uuid.UUID, chapterIDs []string, unlockedAt time.Time) error
	LockPurchasedChapters(ctx context.Context, paymentID uuid.UUID) error
}

type chapterUnlockService struct {
	chapterUnlocks repository.ChapterUnlockRepository
}

func NewChapterUnlockService(chapterUnlocks repository.ChapterUnlockRepository) ChapterUnlockService {
	return &chapterUnlockService{chapterUnlocks: chapterUnlocks}
}

func (unlocks *chapterUnlockService) UnlockPurchasedChapters(ctx context.Context, userID uuid.UUID, paymentID uuid.UUID, chapterIDs []string, unlockedAt time.Time) error {
	return unlocks.chapterUnlocks.UnlockForPayment(ctx, userID, paymentID, chapterIDs, unlockedAt)
}

func (unlocks *chapterUnlockService) LockPurchasedChapters(ctx context.Context, paymentID uuid.UUID) error {
	return unlocks.chapterUnlocks.LockForPayment(ctx, paymentID)
}
