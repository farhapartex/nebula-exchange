package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/authentication"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/request"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
	"github.com/farhapartex/nebula-exchange/backend/internal/story/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/story/service"
)

const levelQueryParameter = "level"

type StorySlideResponse struct {
	ID          string                `json:"id"`
	Position    int                   `json:"position"`
	Kind        models.SlideKind      `json:"kind"`
	Eyebrow     *string               `json:"eyebrow"`
	Heading     string                `json:"heading"`
	Body        string                `json:"body"`
	Image       *string               `json:"image"`
	Palette     database.JSONDocument `json:"palette"`
	ButtonLabel *string               `json:"button_label"`
}

type StoryHandler struct {
	storyService service.StoryService
}

func NewStoryHandler(storyService service.StoryService) *StoryHandler {
	return &StoryHandler{storyService: storyService}
}

func (handler *StoryHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/stories", authentication.RequireUser(), handler.listStories)
}

func (handler *StoryHandler) listStories(context *gin.Context) {
	levelID := context.Query(levelQueryParameter)
	if levelID == "" {
		response.WriteError(context, apierror.ValidationFailed(map[string]string{levelQueryParameter: "is required"}))
		return
	}
	pageRequest, err := request.PaginationFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	var after service.StoryCursor
	if pageRequest.HasCursor() {
		after, err = request.DecodeCursorPosition[service.StoryCursor](pageRequest)
		if err != nil {
			response.WriteError(context, err)
			return
		}
	}

	storyPage, err := handler.storyService.ListLevelStory(context.Request.Context(), levelID, after, pageRequest)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteList(context, http.StatusOK, pagination.Page[StorySlideResponse]{
		Items: toStorySlideResponses(storyPage.Items),
		Info:  storyPage.Info,
	})
}

func toStorySlideResponses(storySlides []service.StorySlide) []StorySlideResponse {
	responses := make([]StorySlideResponse, 0, len(storySlides))
	for _, storySlide := range storySlides {
		responses = append(responses, StorySlideResponse{
			ID:          storySlide.ID,
			Position:    storySlide.Position,
			Kind:        storySlide.Kind,
			Eyebrow:     storySlide.Eyebrow,
			Heading:     storySlide.Heading,
			Body:        storySlide.Body,
			Image:       storySlide.ImageURL,
			Palette:     storySlide.Palette,
			ButtonLabel: storySlide.ButtonLabel,
		})
	}
	return responses
}
