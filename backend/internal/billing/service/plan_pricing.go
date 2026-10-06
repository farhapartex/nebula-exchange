package service

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/farhapartex/nebula-exchange/backend/internal/billing/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

const (
	minimumBundleChapters   = 2
	minimumAllChapters      = 2
	maximumDiscountPercent  = 90
	singleChapterCount      = 1
	percentRoundingDivisor  = 100
	percentRoundingHalfStep = 50
)

type DiscountTier struct {
	MinimumChapters int `json:"min_chapters"`
	Percent         int `json:"percent"`
}

type PricedChapter struct {
	ID         string
	Number     int
	Title      string
	PriceCents int64
}

type PlanOption struct {
	ChapterCount    int
	Chapters        []PricedChapter
	SubtotalCents   int64
	DiscountPercent int
	DiscountCents   int64
	TotalCents      int64
}

func ParseDiscountTiers(document database.JSONDocument) ([]DiscountTier, error) {
	var discountTiers []DiscountTier
	if len(document) == 0 {
		return discountTiers, nil
	}
	if err := json.Unmarshal(document, &discountTiers); err != nil {
		return nil, fmt.Errorf("discount tiers: %w", err)
	}
	for tierIndex, discountTier := range discountTiers {
		if discountTier.MinimumChapters < 1 || discountTier.Percent < 0 || discountTier.Percent > maximumDiscountPercent {
			return nil, fmt.Errorf("discount tier %d needs min_chapters of 1 or more and a percent from 0 to %d", tierIndex+1, maximumDiscountPercent)
		}
		if tierIndex > 0 && discountTier.MinimumChapters <= discountTiers[tierIndex-1].MinimumChapters {
			return nil, fmt.Errorf("discount tier %d must need more chapters than the one before it", tierIndex+1)
		}
	}
	return discountTiers, nil
}

func discountPercentFor(discountTiers []DiscountTier, chapterCount int) int {
	discountPercent := 0
	for _, discountTier := range discountTiers {
		if chapterCount >= discountTier.MinimumChapters {
			discountPercent = discountTier.Percent
		}
	}
	return discountPercent
}

func quoteOption(chapters []PricedChapter, discountTiers []DiscountTier) PlanOption {
	var subtotalCents int64
	for _, chapter := range chapters {
		subtotalCents += chapter.PriceCents
	}
	discountPercent := discountPercentFor(discountTiers, len(chapters))
	discountCents := (subtotalCents*int64(discountPercent) + percentRoundingHalfStep) / percentRoundingDivisor
	return PlanOption{
		ChapterCount:    len(chapters),
		Chapters:        chapters,
		SubtotalCents:   subtotalCents,
		DiscountPercent: discountPercent,
		DiscountCents:   discountCents,
		TotalCents:      subtotalCents - discountCents,
	}
}

func optionsForPlan(planKind models.PlanKind, chaptersToBuy []PricedChapter, discountTiers []DiscountTier) []PlanOption {
	switch planKind {
	case models.PlanKindSingleChapter:
		if len(chaptersToBuy) < singleChapterCount {
			return nil
		}
		return []PlanOption{quoteOption(chaptersToBuy[:singleChapterCount], discountTiers)}
	case models.PlanKindChapterBundle:
		var options []PlanOption
		for chapterCount := minimumBundleChapters; chapterCount < len(chaptersToBuy); chapterCount++ {
			options = append(options, quoteOption(chaptersToBuy[:chapterCount], discountTiers))
		}
		return options
	case models.PlanKindAllChapters:
		if len(chaptersToBuy) < minimumAllChapters {
			return nil
		}
		return []PlanOption{quoteOption(chaptersToBuy, discountTiers)}
	}
	return nil
}

func sortByChapterNumber(chapters []PricedChapter) {
	sort.Slice(chapters, func(first, second int) bool { return chapters[first].Number < chapters[second].Number })
}
