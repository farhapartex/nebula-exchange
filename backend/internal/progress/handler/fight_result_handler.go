package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/authentication"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/request"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/service"
)

type fightResultRequest struct {
	Outcome     string `json:"outcome" binding:"required,oneof=WON LOST"`
	DurationMS  *int   `json:"duration_ms" binding:"required"`
	DamageDealt *int   `json:"damage_dealt" binding:"required"`
	DamageTaken *int   `json:"damage_taken" binding:"required"`
}

type FightResultResponse struct {
	ID               uuid.UUID           `json:"id"`
	Level            string              `json:"level"`
	Status           models.FightStatus  `json:"status"`
	Outcome          models.FightOutcome `json:"outcome"`
	Stars            int                 `json:"stars"`
	DurationMS       int                 `json:"duration_ms"`
	DamageDealt      int                 `json:"damage_dealt"`
	DamageTaken      int                 `json:"damage_taken"`
	IsLevelCompleted bool                `json:"is_level_completed"`
	RewardCoins      string              `json:"reward_coins"`
	RewardExperience int                 `json:"reward_experience"`
	FinishedAt       time.Time           `json:"finished_at"`
}

type FightResultHandler struct {
	fightResultService service.FightResultService
}

func NewFightResultHandler(fightResultService service.FightResultService) *FightResultHandler {
	return &FightResultHandler{fightResultService: fightResultService}
}

func (handler *FightResultHandler) RegisterRoutes(router gin.IRouter) {
	router.POST("/fight-sessions/:fightSessionID/results", authentication.RequireUser(), handler.postFightResult)
}

func (handler *FightResultHandler) postFightResult(context *gin.Context) {
	fightSessionID, err := uuid.Parse(context.Param("fightSessionID"))
	if err != nil {
		response.WriteError(context, service.ErrFightNotFound)
		return
	}
	var body fightResultRequest
	if err := request.BindJSON(context, &body); err != nil {
		response.WriteError(context, err)
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	fightResult, err := handler.fightResultService.Submit(context.Request.Context(), userID, fightSessionID, service.FightReport{
		Outcome:     models.FightOutcome(body.Outcome),
		DurationMS:  *body.DurationMS,
		DamageDealt: *body.DamageDealt,
		DamageTaken: *body.DamageTaken,
	})
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusCreated, FightResultResponse{
		ID:               fightResult.FightSessionID,
		Level:            fightResult.LevelID,
		Status:           fightResult.Status,
		Outcome:          fightResult.Outcome,
		Stars:            fightResult.Stars,
		DurationMS:       fightResult.DurationMS,
		DamageDealt:      fightResult.DamageDealt,
		DamageTaken:      fightResult.DamageTaken,
		IsLevelCompleted: fightResult.IsLevelCompleted,
		RewardCoins:      "0",
		RewardExperience: 0,
		FinishedAt:       fightResult.FinishedAt,
	})
}
