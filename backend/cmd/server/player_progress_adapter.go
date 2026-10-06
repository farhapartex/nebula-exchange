package main

import (
	"context"

	"github.com/google/uuid"

	identityservice "github.com/farhapartex/nebula-exchange/backend/internal/identity/service"
	progressservice "github.com/farhapartex/nebula-exchange/backend/internal/progress/service"
)

type playerProgressAdapter struct {
	playerProgress progressservice.PlayerProgressService
}

func (adapter playerProgressAdapter) PlayerProgressOf(ctx context.Context, userID uuid.UUID) (identityservice.PlayerProgress, error) {
	playerProgress, err := adapter.playerProgress.PlayerProgress(ctx, userID)
	if err != nil {
		return identityservice.PlayerProgress{}, err
	}
	return identityservice.PlayerProgress{
		CurrentLevel:       playerProgress.CurrentLevel,
		StoryLevel:         playerProgress.StoryLevel,
		TotalWins:          playerProgress.TotalWins,
		TotalLosses:        playerProgress.TotalLosses,
		CurrentLevelWins:   playerProgress.CurrentLevelWins,
		CurrentLevelLosses: playerProgress.CurrentLevelLosses,
	}, nil
}
