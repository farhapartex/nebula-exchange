package service

import (
	"context"

	"github.com/google/uuid"

	storyservice "github.com/farhapartex/nebula-exchange/backend/internal/story/service"
)

type ChapterCatalog interface {
	ChaptersOnSale(ctx context.Context) ([]storyservice.ChapterForSale, error)
}

type ChapterOwnership interface {
	OwnedChapterIDs(ctx context.Context, userID uuid.UUID) (map[string]bool, error)
}
