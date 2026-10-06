package service

import (
	"context"
	"time"

	storyservice "github.com/farhapartex/nebula-exchange/backend/internal/story/service"
)

type LevelCatalog interface {
	PlayableLevel(ctx context.Context, levelID string) (storyservice.LevelPlacement, error)
	Placements(ctx context.Context) (map[string]storyservice.LevelPlacement, error)
}

type TransactionRunner interface {
	WithinTransaction(ctx context.Context, work func(ctx context.Context) error) error
}

type Clock func() time.Time
