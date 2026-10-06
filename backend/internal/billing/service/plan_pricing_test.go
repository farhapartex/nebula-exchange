package service

import (
	"testing"

	"github.com/farhapartex/nebula-exchange/backend/internal/billing/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

func chaptersNumbered(count int) []PricedChapter {
	chapters := make([]PricedChapter, 0, count)
	for chapterNumber := 1; chapterNumber <= count; chapterNumber++ {
		chapters = append(chapters, PricedChapter{ID: string(rune('0' + chapterNumber)), Number: chapterNumber, PriceCents: 499})
	}
	return chapters
}

func mustParseTiers(t *testing.T, document string) []DiscountTier {
	t.Helper()
	discountTiers, err := ParseDiscountTiers(database.JSONDocument(document))
	if err != nil {
		t.Fatalf("parse tiers: %v", err)
	}
	return discountTiers
}

func TestSingleChapterPlanQuotesTheNextChapterAtFullPrice(t *testing.T) {
	options := optionsForPlan(models.PlanKindSingleChapter, chaptersNumbered(5), nil)
	if len(options) != 1 || options[0].ChapterCount != 1 || options[0].TotalCents != 499 || options[0].Chapters[0].Number != 1 {
		t.Fatalf("unexpected single chapter options %+v", options)
	}
}

func TestBundlePlanOffersEveryCountBelowAllWithTieredDiscounts(t *testing.T) {
	bundleTiers := mustParseTiers(t, `[{"min_chapters":2,"percent":5},{"min_chapters":3,"percent":10}]`)
	options := optionsForPlan(models.PlanKindChapterBundle, chaptersNumbered(5), bundleTiers)
	if len(options) != 3 {
		t.Fatalf("got %d bundle options, want counts 2, 3 and 4", len(options))
	}
	expectedTotals := map[int]int64{2: 948, 3: 1347, 4: 1796}
	expectedPercents := map[int]int{2: 5, 3: 10, 4: 10}
	for _, option := range options {
		if option.TotalCents != expectedTotals[option.ChapterCount] || option.DiscountPercent != expectedPercents[option.ChapterCount] {
			t.Fatalf("count %d: got %+v", option.ChapterCount, option)
		}
	}
}

func TestAllChaptersPlanCoversEveryRemainingChapterAndNeedsTwo(t *testing.T) {
	allTiers := mustParseTiers(t, `[{"min_chapters":2,"percent":20}]`)
	options := optionsForPlan(models.PlanKindAllChapters, chaptersNumbered(5), allTiers)
	if len(options) != 1 || options[0].ChapterCount != 5 || options[0].SubtotalCents != 2495 || options[0].TotalCents != 1996 {
		t.Fatalf("unexpected all chapters option %+v", options)
	}
	if optionsForPlan(models.PlanKindAllChapters, chaptersNumbered(1), allTiers) != nil {
		t.Fatal("all chapters with one chapter left is the same as the single plan and must not be offered")
	}
	if optionsForPlan(models.PlanKindChapterBundle, chaptersNumbered(2), allTiers) != nil {
		t.Fatal("a bundle needs more chapters left than the bundle itself")
	}
}

func TestBrokenDiscountTiersAreRejected(t *testing.T) {
	for _, brokenTiers := range []string{`[{"min_chapters":0,"percent":5}]`, `[{"min_chapters":2,"percent":95}]`, `[{"min_chapters":3,"percent":5},{"min_chapters":2,"percent":10}]`, `{"min_chapters":2}`} {
		if _, err := ParseDiscountTiers(database.JSONDocument(brokenTiers)); err == nil {
			t.Fatalf("expected %s to be rejected", brokenTiers)
		}
	}
}
