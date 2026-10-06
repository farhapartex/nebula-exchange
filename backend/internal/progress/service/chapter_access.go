package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/progress/repository"
	storyservice "github.com/farhapartex/nebula-exchange/backend/internal/story/service"
)

type ownedChapters map[string]bool

func loadOwnedChapters(ctx context.Context, chapterUnlocks repository.ChapterUnlockRepository, userID uuid.UUID) (ownedChapters, error) {
	unlockedChapterIDs, err := chapterUnlocks.ListUnlockedChapterIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	owned := make(ownedChapters, len(unlockedChapterIDs))
	for _, chapterID := range unlockedChapterIDs {
		owned[chapterID] = true
	}
	return owned, nil
}

func (owned ownedChapters) hasPaidFor(placement storyservice.LevelPlacement) bool {
	return placement.IsChapterFree || owned[placement.ChapterID]
}

func (owned ownedChapters) canPlay(placement storyservice.LevelPlacement) bool {
	return placement.IsLevelFree || owned.hasPaidFor(placement)
}
