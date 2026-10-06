package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/authentication"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/service"
)

type NextChapterResponse struct {
	Number int     `json:"number"`
	Title  *string `json:"title"`
}

type NextLevelDetailsResponse struct {
	ID               string `json:"id"`
	Number           int    `json:"number"`
	Title            string `json:"title"`
	Teaser           string `json:"teaser"`
	TimeLimitSeconds int    `json:"time_limit_seconds"`
	BestStars        *int   `json:"best_stars"`
	Attempts         int    `json:"attempts"`
}

type NextLevelResponse struct {
	Status  service.NextLevelStatus   `json:"status"`
	Chapter NextChapterResponse       `json:"chapter"`
	Level   *NextLevelDetailsResponse `json:"level"`
}

type NextLevelHandler struct {
	playerProgressService service.PlayerProgressService
}

func NewNextLevelHandler(playerProgressService service.PlayerProgressService) *NextLevelHandler {
	return &NextLevelHandler{playerProgressService: playerProgressService}
}

func (handler *NextLevelHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/me/next-level", authentication.RequireUser(), handler.getNextLevel)
}

func (handler *NextLevelHandler) getNextLevel(context *gin.Context) {
	userID, _ := authentication.UserIDFrom(context)
	playerProgress, err := handler.playerProgressService.PlayerProgress(context.Request.Context(), userID)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	nextLevel := playerProgress.NextLevel
	nextLevelResponse := NextLevelResponse{
		Status:  nextLevel.Status,
		Chapter: NextChapterResponse{Number: nextLevel.ChapterNumber, Title: nextLevel.ChapterTitle},
	}
	if nextLevel.Level != nil {
		nextLevelResponse.Level = &NextLevelDetailsResponse{
			ID:               nextLevel.Level.LevelID,
			Number:           nextLevel.Level.LevelNumber,
			Title:            nextLevel.Level.Title,
			Teaser:           nextLevel.Level.Teaser,
			TimeLimitSeconds: nextLevel.Level.TimeLimitSeconds,
			BestStars:        nextLevel.BestStars,
			Attempts:         nextLevel.Attempts,
		}
	}
	response.WriteData(context, http.StatusOK, nextLevelResponse)
}
