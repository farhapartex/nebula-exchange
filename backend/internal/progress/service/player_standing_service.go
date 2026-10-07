package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/repository"
)

type PlayerStandingService interface {
	FighterLevel(ctx context.Context, userID uuid.UUID) (int, error)
	WonLevelIDs(ctx context.Context, userID uuid.UUID) (map[string]bool, error)
}

type playerStandingService struct {
	playerProgress PlayerProgressService
	levelProgress  repository.LevelProgressRepository
}

func NewPlayerStandingService(playerProgress PlayerProgressService, levelProgress repository.LevelProgressRepository) PlayerStandingService {
	return &playerStandingService{playerProgress: playerProgress, levelProgress: levelProgress}
}

func (standing *playerStandingService) FighterLevel(ctx context.Context, userID uuid.UUID) (int, error) {
	progress, err := standing.playerProgress.PlayerProgress(ctx, userID)
	if err != nil {
		return 0, err
	}
	return progress.CurrentLevel, nil
}

func (standing *playerStandingService) WonLevelIDs(ctx context.Context, userID uuid.UUID) (map[string]bool, error) {
	levelProgressRows, err := standing.levelProgress.ListForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	wonLevelIDs := make(map[string]bool, len(levelProgressRows))
	for _, levelProgressRow := range levelProgressRows {
		if levelProgressRow.Status == models.ProgressStatusCompleted {
			wonLevelIDs[levelProgressRow.LevelID] = true
		}
	}
	return wonLevelIDs, nil
}
