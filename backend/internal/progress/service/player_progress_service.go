package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/progress/repository"
)

type PlayerProgress struct {
	FighterLevel int
	StoryLevel   *string
	Wins         int
	Losses       int
}

type PlayerProgressService interface {
	PlayerProgress(ctx context.Context, userID uuid.UUID) (PlayerProgress, error)
}

type playerProgressService struct {
	levels        LevelCatalog
	levelProgress repository.LevelProgressRepository
	fighters      fighterResolver
}

func NewPlayerProgressService(levels LevelCatalog, levelProgress repository.LevelProgressRepository, fighters repository.FighterRepository) PlayerProgressService {
	return &playerProgressService{levels: levels, levelProgress: levelProgress, fighters: fighterResolver{fighters: fighters}}
}

func (progress *playerProgressService) PlayerProgress(ctx context.Context, userID uuid.UUID) (PlayerProgress, error) {
	storyLevel, err := progress.furthestStartedLevel(ctx, userID)
	if err != nil {
		return PlayerProgress{}, err
	}
	fighter, err := progress.fighters.currentFighter(ctx, userID)
	if errors.Is(err, ErrGameDataMissing) {
		return PlayerProgress{StoryLevel: storyLevel}, nil
	}
	if err != nil {
		return PlayerProgress{}, err
	}
	return PlayerProgress{FighterLevel: fighter.Level, StoryLevel: storyLevel, Wins: fighter.Wins, Losses: fighter.Loss}, nil
}

func (progress *playerProgressService) furthestStartedLevel(ctx context.Context, userID uuid.UUID) (*string, error) {
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
