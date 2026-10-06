package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity/service"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/authentication"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
)

type CurrentPlayerResponse struct {
	Name             string `json:"name"`
	Email            string `json:"email"`
	CurrentLevel     int    `json:"current_level"`
	StoryLevel       int    `json:"story_level"`
	TotalWin         int    `json:"total_win"`
	TotalLose        int    `json:"total_lose"`
	CurrentLevelWin  int    `json:"current_level_win"`
	CurrentLevelLose int    `json:"current_level_lose"`
}

type ProfileHandler struct {
	profileService service.ProfileService
}

func NewProfileHandler(profileService service.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

func (handler *ProfileHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/me", authentication.RequireUser(), handler.getMe)
}

func (handler *ProfileHandler) getMe(context *gin.Context) {
	userID, _ := authentication.UserIDFrom(context)
	currentPlayer, err := handler.profileService.CurrentPlayer(context.Request.Context(), userID)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, CurrentPlayerResponse{
		Name:             currentPlayer.Name,
		Email:            currentPlayer.Email,
		CurrentLevel:     currentPlayer.CurrentLevel,
		StoryLevel:       currentPlayer.StoryLevel,
		TotalWin:         currentPlayer.TotalWins,
		TotalLose:        currentPlayer.TotalLosses,
		CurrentLevelWin:  currentPlayer.CurrentLevelWins,
		CurrentLevelLose: currentPlayer.CurrentLevelLosses,
	})
}
