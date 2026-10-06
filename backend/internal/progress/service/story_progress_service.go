package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/progress/repository"
)

type StoryProgressService interface {
	FurthestStartedLevel(ctx context.Context, userID uuid.UUID) (*string, error)
}

type storyProgressService struct {
	levels        LevelCatalog
	levelProgress repository.LevelProgressRepository
}

func NewStoryProgressService(levels LevelCatalog, levelProgress repository.LevelProgressRepository) StoryProgressService {
	return &storyProgressService{levels: levels, levelProgress: levelProgress}
}

func (progress *storyProgressService) FurthestStartedLevel(ctx context.Context, userID uuid.UUID) (*string, error) {
	startedLevelIDs, err := progress.levelProgress.ListStartedLevelIDs(ctx, userID)
	if err != nil || len(startedLevelIDs) == 0 {
		return nil, err
	}
	placements, err := progress.levels.Placements(ctx)
	if err != nil {
		return nil, err
	}
	var furthestLevelID *string
	furthestOrder := -1
	for _, startedLevelID := range startedLevelIDs {
		placement, isKnown := placements[startedLevelID]
		if isKnown && placement.Order > furthestOrder {
			furthestOrder = placement.Order
			levelID := startedLevelID
			furthestLevelID = &levelID
		}
	}
	return furthestLevelID, nil
}
