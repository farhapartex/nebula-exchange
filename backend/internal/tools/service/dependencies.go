package service

import (
	"context"

	"github.com/google/uuid"

	storyservice "github.com/farhapartex/nebula-exchange/backend/internal/story/service"
)

type PlayerStanding interface {
	FighterLevel(ctx context.Context, userID uuid.UUID) (int, error)
	WonLevelIDs(ctx context.Context, userID uuid.UUID) (map[string]bool, error)
}

type LevelDirectory interface {
	Placements(ctx context.Context) (map[string]storyservice.LevelPlacement, error)
}
