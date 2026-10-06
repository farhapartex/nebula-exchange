package service

import (
	"context"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/story/repository"
)

var ErrLevelNotPlayable = apierror.NotFound("This level does not exist or cannot be played yet")

type LevelPlacement struct {
	LevelID         string
	ChapterNumber   int
	LevelNumber     int
	Order           int
	PreviousLevelID *string
}

type LevelCatalog interface {
	PlayableLevel(ctx context.Context, levelID string) (LevelPlacement, error)
	Placements(ctx context.Context) (map[string]LevelPlacement, error)
	FightContent(ctx context.Context, levelID string) (FightContent, error)
}

type levelCatalog struct {
	levels repository.LevelRepository
}

func NewLevelCatalog(levels repository.LevelRepository) LevelCatalog {
	return &levelCatalog{levels: levels}
}

func (catalog *levelCatalog) PlayableLevel(ctx context.Context, levelID string) (LevelPlacement, error) {
	placements, err := catalog.Placements(ctx)
	if err != nil {
		return LevelPlacement{}, err
	}
	placement, isPlayable := placements[levelID]
	if !isPlayable {
		return LevelPlacement{}, ErrLevelNotPlayable
	}
	return placement, nil
}

func (catalog *levelCatalog) Placements(ctx context.Context) (map[string]LevelPlacement, error) {
	orderedLevels, err := catalog.levels.ListPublishedStoryLevels(ctx)
	if err != nil {
		return nil, err
	}
	placements := make(map[string]LevelPlacement, len(orderedLevels))
	var previousLevelID *string
	for levelOrder, level := range orderedLevels {
		placements[level.ID] = LevelPlacement{
			LevelID:         level.ID,
			ChapterNumber:   level.Chapter.Number,
			LevelNumber:     *level.Number,
			Order:           levelOrder,
			PreviousLevelID: previousLevelID,
		}
		currentLevelID := level.ID
		previousLevelID = &currentLevelID
	}
	return placements, nil
}
