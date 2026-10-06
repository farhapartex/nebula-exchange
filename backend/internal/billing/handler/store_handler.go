package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/billing/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/billing/service"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/authentication"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/request"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
)

type ChapterResponse struct {
	ID         string  `json:"id"`
	Number     int     `json:"number"`
	Title      string  `json:"title"`
	IsFree     bool    `json:"is_free"`
	PriceCents *string `json:"price_cents"`
	LevelCount int     `json:"level_count"`
	IsOwned    bool    `json:"is_owned"`
}

type PlanChapterResponse struct {
	ID         string `json:"id"`
	Number     int    `json:"number"`
	Title      string `json:"title"`
	PriceCents string `json:"price_cents"`
}

type PlanOptionResponse struct {
	ChapterCount    int                   `json:"chapter_count"`
	Chapters        []PlanChapterResponse `json:"chapters"`
	SubtotalCents   string                `json:"subtotal_cents"`
	DiscountPercent int                   `json:"discount_percent"`
	DiscountCents   string                `json:"discount_cents"`
	TotalCents      string                `json:"total_cents"`
}

type PlanResponse struct {
	ID          string               `json:"id"`
	Kind        models.PlanKind      `json:"kind"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	IsAvailable bool                 `json:"is_available"`
	Options     []PlanOptionResponse `json:"options"`
}

type StoreHandler struct {
	storeService service.StoreService
}

func NewStoreHandler(storeService service.StoreService) *StoreHandler {
	return &StoreHandler{storeService: storeService}
}

func (handler *StoreHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/chapters", authentication.RequireUser(), handler.listChapters)
	router.GET("/plans", authentication.RequireUser(), handler.listPlans)
}

func (handler *StoreHandler) listChapters(context *gin.Context) {
	pageRequest, err := request.PaginationFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	var after service.ChapterCursor
	if pageRequest.HasCursor() {
		if after, err = request.DecodeCursorPosition[service.ChapterCursor](pageRequest); err != nil {
			response.WriteError(context, err)
			return
		}
	}
	userID, _ := authentication.UserIDFrom(context)
	chapterPage, err := handler.storeService.ListChapters(context.Request.Context(), userID, after, pageRequest)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	chapterResponses := make([]ChapterResponse, 0, len(chapterPage.Items))
	for _, chapter := range chapterPage.Items {
		chapterResponses = append(chapterResponses, ChapterResponse{
			ID:         chapter.ID,
			Number:     chapter.Number,
			Title:      chapter.Title,
			IsFree:     chapter.IsFree,
			PriceCents: formatOptionalCents(chapter.PriceCents),
			LevelCount: chapter.LevelCount,
			IsOwned:    chapter.IsOwned,
		})
	}
	response.WriteList(context, http.StatusOK, pagination.Page[ChapterResponse]{Items: chapterResponses, Info: chapterPage.Info})
}

func (handler *StoreHandler) listPlans(context *gin.Context) {
	pageRequest, err := request.PaginationFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	var after service.PlanCursor
	if pageRequest.HasCursor() {
		if after, err = request.DecodeCursorPosition[service.PlanCursor](pageRequest); err != nil {
			response.WriteError(context, err)
			return
		}
	}
	userID, _ := authentication.UserIDFrom(context)
	planPage, err := handler.storeService.ListPlans(context.Request.Context(), userID, after, pageRequest)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	planResponses := make([]PlanResponse, 0, len(planPage.Items))
	for _, planOffer := range planPage.Items {
		planResponses = append(planResponses, toPlanResponse(planOffer))
	}
	response.WriteList(context, http.StatusOK, pagination.Page[PlanResponse]{Items: planResponses, Info: planPage.Info})
}

func toPlanResponse(planOffer service.PlanOffer) PlanResponse {
	optionResponses := make([]PlanOptionResponse, 0, len(planOffer.Options))
	for _, option := range planOffer.Options {
		chapterResponses := make([]PlanChapterResponse, 0, len(option.Chapters))
		for _, chapter := range option.Chapters {
			chapterResponses = append(chapterResponses, PlanChapterResponse{ID: chapter.ID, Number: chapter.Number, Title: chapter.Title, PriceCents: formatCents(chapter.PriceCents)})
		}
		optionResponses = append(optionResponses, PlanOptionResponse{
			ChapterCount:    option.ChapterCount,
			Chapters:        chapterResponses,
			SubtotalCents:   formatCents(option.SubtotalCents),
			DiscountPercent: option.DiscountPercent,
			DiscountCents:   formatCents(option.DiscountCents),
			TotalCents:      formatCents(option.TotalCents),
		})
	}
	return PlanResponse{
		ID:          planOffer.ID,
		Kind:        planOffer.Kind,
		Name:        planOffer.Name,
		Description: planOffer.Description,
		IsAvailable: planOffer.IsAvailable,
		Options:     optionResponses,
	}
}
