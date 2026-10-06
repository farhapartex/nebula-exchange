package service

import (
	"context"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/story/repository"
)

var ErrLevelNotPlayable = apierror.NotFound("This level does not exist or cannot be played yet")

type LevelPlacement struct {
	LevelID          string
	ChapterID        string
	ChapterNumber    int
	ChapterTitle     string
	IsChapterFree    bool
	ChapterPrice     *int64
	IsLevelFree      bool
	LevelNumber      int
	Title            string
	Teaser           string
	TimeLimitSeconds int
	Order            int
	PreviousLevelID  *string
}

type LevelCatalog interface {
	PlayableLevel(ctx context.Context, levelID string) (LevelPlacement, error)
	Placements(ctx context.Context) (map[string]LevelPlacement, error)
	OrderedPlacements(ctx context.Context) ([]LevelPlacement, error)
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
	orderedPlacements, err := catalog.OrderedPlacements(ctx)
	if err != nil {
		return nil, err
	}
	placements := make(map[string]LevelPlacement, len(orderedPlacements))
	for _, placement := range orderedPlacements {
		placements[placement.LevelID] = placement
	}
	return placements, nil
}

func (catalog *levelCatalog) OrderedPlacements(ctx context.Context) ([]LevelPlacement, error) {
	orderedLevels, err := catalog.levels.ListPublishedStoryLevels(ctx)
	if err != nil {
		return nil, err
	}
	orderedPlacements := make([]LevelPlacement, 0, len(orderedLevels))
	var previousLevelID *string
	for levelOrder, level := range orderedLevels {
		orderedPlacements = append(orderedPlacements, LevelPlacement{
			LevelID:          level.ID,
			ChapterID:        level.Chapter.ID,
			ChapterNumber:    level.Chapter.Number,
			ChapterTitle:     level.Chapter.Title,
			IsChapterFree:    level.Chapter.IsFree,
			ChapterPrice:     level.Chapter.PriceCoins,
			IsLevelFree:      level.IsFree,
			LevelNumber:      *level.Number,
			Title:            level.Title,
			Teaser:           level.Teaser,
			TimeLimitSeconds: level.TimeLimitSeconds,
			Order:            levelOrder,
			PreviousLevelID:  previousLevelID,
		})
		currentLevelID := level.ID
		previousLevelID = &currentLevelID
	}
	return orderedPlacements, nil
}
