package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/authentication"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/service"
)

type ArenaResponse struct {
	Name   string                `json:"name"`
	Width  int                   `json:"width"`
	FloorY int                   `json:"floor_y"`
	Stage  database.JSONDocument `json:"stage"`
}

type PlayerResponse struct {
	ID    string                `json:"id"`
	Name  string                `json:"name"`
	Title string                `json:"title"`
	Stats database.JSONDocument `json:"stats"`
	Look  database.JSONDocument `json:"look"`
}

type EnemyResponse struct {
	ID    string                `json:"id"`
	Name  string                `json:"name"`
	Title string                `json:"title"`
	Stats database.JSONDocument `json:"stats"`
	Look  database.JSONDocument `json:"look"`
	Brain database.JSONDocument `json:"brain"`
}

type FightSetupResponse struct {
	LevelID          string         `json:"level_id"`
	TimeLimitSeconds int            `json:"time_limit_seconds"`
	Arena            ArenaResponse  `json:"arena"`
	Player           PlayerResponse `json:"player"`
	Enemy            EnemyResponse  `json:"enemy"`
}

type FightSetupHandler struct {
	fightSetupService service.FightSetupService
}

func NewFightSetupHandler(fightSetupService service.FightSetupService) *FightSetupHandler {
	return &FightSetupHandler{fightSetupService: fightSetupService}
}

func (handler *FightSetupHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/levels/:levelID/fight-setup", authentication.RequireUser(), handler.getFightSetup)
}

func (handler *FightSetupHandler) getFightSetup(context *gin.Context) {
	userID, _ := authentication.UserIDFrom(context)
	fightSetup, err := handler.fightSetupService.FightSetup(context.Request.Context(), userID, context.Param("levelID"))
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, FightSetupResponse{
		LevelID:          fightSetup.LevelID,
		TimeLimitSeconds: fightSetup.TimeLimitSeconds,
		Arena:            ArenaResponse{Name: fightSetup.Arena.Name, Width: fightSetup.Arena.Width, FloorY: fightSetup.Arena.FloorY, Stage: fightSetup.Arena.Stage},
		Player:           PlayerResponse{ID: fightSetup.Player.ID, Name: fightSetup.Player.Name, Title: fightSetup.Player.Title, Stats: fightSetup.Player.Stats, Look: fightSetup.Player.Look},
		Enemy:            EnemyResponse{ID: fightSetup.Enemy.ID, Name: fightSetup.Enemy.Name, Title: fightSetup.Enemy.Title, Stats: fightSetup.Enemy.Stats, Look: fightSetup.Enemy.Look, Brain: fightSetup.Enemy.Brain},
	})
}
