package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
)

type ChapterListing struct {
	ID         string
	Number     int
	Title      string
	IsFree     bool
	PriceCents *int64
	LevelCount int
	IsOwned    bool
}

type PlanOffer struct {
	ID          string
	Kind        models.PlanKind
	Name        string
	Description string
	SortOrder   int
	IsAvailable bool
	Options     []PlanOption
}

type ChapterCursor struct {
	Number int `json:"number"`
}

type PlanCursor struct {
	SortOrder int `json:"sort_order"`
}

type StoreService interface {
	ListChapters(ctx context.Context, userID uuid.UUID, after ChapterCursor, pageRequest pagination.Request) (pagination.Page[ChapterListing], error)
	ListPlans(ctx context.Context, userID uuid.UUID, after PlanCursor, pageRequest pagination.Request) (pagination.Page[PlanOffer], error)
}

type storeService struct {
	quoter chapterQuoter
	plans  repository.PlanRepository
}

func NewStoreService(chapters ChapterCatalog, ownership ChapterOwnership, plans repository.PlanRepository) StoreService {
	return &storeService{quoter: chapterQuoter{chapters: chapters, ownership: ownership}, plans: plans}
}

func (store *storeService) ListChapters(ctx context.Context, userID uuid.UUID, after ChapterCursor, pageRequest pagination.Request) (pagination.Page[ChapterListing], error) {
	chapterListings, err := store.quoter.chapterListings(ctx, userID)
	if err != nil {
		return pagination.Page[ChapterListing]{}, err
	}
	var pageCandidates []ChapterListing
	for _, chapterListing := range chapterListings {
		if chapterListing.Number > after.Number && len(pageCandidates) < pageRequest.FetchLimit() {
			pageCandidates = append(pageCandidates, chapterListing)
		}
	}
	return pagination.BuildPage(pageCandidates, pageRequest, func(chapterListing ChapterListing) ChapterCursor {
		return ChapterCursor{Number: chapterListing.Number}
	})
}

func (store *storeService) ListPlans(ctx context.Context, userID uuid.UUID, after PlanCursor, pageRequest pagination.Request) (pagination.Page[PlanOffer], error) {
	plans, err := store.plans.ListActive(ctx, after.SortOrder, pageRequest.FetchLimit())
	if err != nil {
		return pagination.Page[PlanOffer]{}, err
	}
	chapterListings, err := store.quoter.chapterListings(ctx, userID)
	if err != nil {
		return pagination.Page[PlanOffer]{}, err
	}
	chaptersToBuy := chaptersStillToBuy(chapterListings)

	planOffers := make([]PlanOffer, 0, len(plans))
	for _, plan := range plans {
		discountTiers, err := ParseDiscountTiers(plan.DiscountTiers)
		if err != nil {
			return pagination.Page[PlanOffer]{}, err
		}
		options := optionsForPlan(plan.Kind, chaptersToBuy, discountTiers)
		planOffers = append(planOffers, PlanOffer{
			ID:          plan.ID,
			Kind:        plan.Kind,
			Name:        plan.Name,
			Description: plan.Description,
			SortOrder:   plan.SortOrder,
			IsAvailable: len(options) > 0,
			Options:     options,
		})
	}
	return pagination.BuildPage(planOffers, pageRequest, func(planOffer PlanOffer) PlanCursor {
		return PlanCursor{SortOrder: planOffer.SortOrder}
	})
}

type chapterQuoter struct {
	chapters  ChapterCatalog
	ownership ChapterOwnership
}

func (quoter chapterQuoter) chaptersToBuy(ctx context.Context, userID uuid.UUID) ([]PricedChapter, error) {
	chapterListings, err := quoter.chapterListings(ctx, userID)
	if err != nil {
		return nil, err
	}
	return chaptersStillToBuy(chapterListings), nil
}

func (quoter chapterQuoter) chapterListings(ctx context.Context, userID uuid.UUID) ([]ChapterListing, error) {
	chaptersOnSale, err := quoter.chapters.ChaptersOnSale(ctx)
	if err != nil {
		return nil, err
	}
	ownedChapterIDs, err := quoter.ownership.OwnedChapterIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	chapterListings := make([]ChapterListing, 0, len(chaptersOnSale))
	for _, chapter := range chaptersOnSale {
		chapterListings = append(chapterListings, ChapterListing{
			ID:         chapter.ID,
			Number:     chapter.Number,
			Title:      chapter.Title,
			IsFree:     chapter.IsFree,
			PriceCents: chapter.PriceCents,
			LevelCount: chapter.LevelCount,
			IsOwned:    ownedChapterIDs[chapter.ID],
		})
	}
	return chapterListings, nil
}

func chaptersStillToBuy(chapterListings []ChapterListing) []PricedChapter {
	var chaptersToBuy []PricedChapter
	for _, chapterListing := range chapterListings {
		if chapterListing.IsFree || chapterListing.IsOwned || chapterListing.PriceCents == nil {
			continue
		}
		chaptersToBuy = append(chaptersToBuy, PricedChapter{
			ID:         chapterListing.ID,
			Number:     chapterListing.Number,
			Title:      chapterListing.Title,
			PriceCents: *chapterListing.PriceCents,
		})
	}
	sortByChapterNumber(chaptersToBuy)
	return chaptersToBuy
}
