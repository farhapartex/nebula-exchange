package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/progress/repository"
)

type ChapterOwnershipService interface {
	OwnedChapterIDs(ctx context.Context, userID uuid.UUID) (map[string]bool, error)
}

type chapterOwnershipService struct {
	chapterUnlocks repository.ChapterUnlockRepository
}

func NewChapterOwnershipService(chapterUnlocks repository.ChapterUnlockRepository) ChapterOwnershipService {
	return &chapterOwnershipService{chapterUnlocks: chapterUnlocks}
}

func (ownership *chapterOwnershipService) OwnedChapterIDs(ctx context.Context, userID uuid.UUID) (map[string]bool, error) {
	return loadOwnedChapters(ctx, ownership.chapterUnlocks, userID)
}
