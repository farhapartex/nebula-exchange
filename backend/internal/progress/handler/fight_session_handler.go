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

type startFightRequest struct {
	Level string `json:"level" binding:"required,max=64"`
}

type FightSessionResponse struct {
	ID        uuid.UUID          `json:"id"`
	Level     string             `json:"level"`
	Status    models.FightStatus `json:"status"`
	Seed      string             `json:"seed"`
	StartedAt time.Time          `json:"started_at"`
}

type FightSessionHandler struct {
	fightSessionService service.FightSessionService
}

func NewFightSessionHandler(fightSessionService service.FightSessionService) *FightSessionHandler {
	return &FightSessionHandler{fightSessionService: fightSessionService}
}

func (handler *FightSessionHandler) RegisterRoutes(router gin.IRouter) {
	router.POST("/fight-sessions", authentication.RequireUser(), handler.postFightSession)
}

func (handler *FightSessionHandler) postFightSession(context *gin.Context) {
	var body startFightRequest
	if err := request.BindJSON(context, &body); err != nil {
		response.WriteError(context, err)
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	startedFight, err := handler.fightSessionService.Start(context.Request.Context(), userID, body.Level)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusCreated, FightSessionResponse{
		ID:        startedFight.ID,
		Level:     startedFight.LevelID,
		Status:    startedFight.Status,
		Seed:      formatSeed(startedFight.Seed),
		StartedAt: startedFight.StartedAt,
	})
}
