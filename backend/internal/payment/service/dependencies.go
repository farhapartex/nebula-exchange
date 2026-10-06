package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	storyservice "github.com/farhapartex/nebula-exchange/backend/internal/story/service"
)

type ChapterCatalog interface {
	ChaptersOnSale(ctx context.Context) ([]storyservice.ChapterForSale, error)
}

type ChapterOwnership interface {
	OwnedChapterIDs(ctx context.Context, userID uuid.UUID) (map[string]bool, error)
}

type ChapterUnlocker interface {
	UnlockPurchasedChapters(ctx context.Context, userID uuid.UUID, paymentID uuid.UUID, chapterIDs []string, unlockedAt time.Time) error
	LockPurchasedChapters(ctx context.Context, paymentID uuid.UUID) error
}

type TransactionRunner interface {
	WithinTransaction(ctx context.Context, work func(ctx context.Context) error) error
}

type Clock func() time.Time
