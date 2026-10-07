package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/authentication"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/request"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/service"
)

var errUnsupportedOutcomeFilter = apierror.ValidationFailed(map[string]string{"outcome": "must be WON"})

type WonFightLevelResponse struct {
	ID            string `json:"id"`
	Number        int    `json:"number"`
	Title         string `json:"title"`
	ChapterNumber int    `json:"chapter_number"`
	ChapterTitle  string `json:"chapter_title"`
}

type WonFightResponse struct {
	ID          uuid.UUID             `json:"id"`
	Level       WonFightLevelResponse `json:"level"`
	Stars       *int                  `json:"stars"`
	DurationMS  int                   `json:"duration_ms"`
	DamageDealt int                   `json:"damage_dealt"`
	DamageTaken int                   `json:"damage_taken"`
	FinishedAt  time.Time             `json:"finished_at"`
}

type WonFightHandler struct {
	wonFightService service.WonFightService
}

func NewWonFightHandler(wonFightService service.WonFightService) *WonFightHandler {
	return &WonFightHandler{wonFightService: wonFightService}
}

func (handler *WonFightHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/fight-sessions", authentication.RequireUser(), handler.listFightSessions)
}

func (handler *WonFightHandler) listFightSessions(context *gin.Context) {
	if models.FightOutcome(context.Query("outcome")) != models.FightOutcomeWon {
		response.WriteError(context, errUnsupportedOutcomeFilter)
		return
	}
	pageRequest, err := request.PaginationFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	var after *service.WonFightCursor
	if pageRequest.HasCursor() {
		cursor, err := request.DecodeCursorPosition[service.WonFightCursor](pageRequest)
		if err != nil {
			response.WriteError(context, err)
			return
		}
		after = &cursor
	}
	userID, _ := authentication.UserIDFrom(context)
	wonFightPage, err := handler.wonFightService.List(context.Request.Context(), userID, after, pageRequest)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	wonFightResponses := make([]WonFightResponse, 0, len(wonFightPage.Items))
	for _, wonFight := range wonFightPage.Items {
		wonFightResponses = append(wonFightResponses, WonFightResponse{
			ID: wonFight.ID,
			Level: WonFightLevelResponse{
				ID:            wonFight.Level.ID,
				Number:        wonFight.Level.Number,
				Title:         wonFight.Level.Title,
				ChapterNumber: wonFight.Level.ChapterNumber,
				ChapterTitle:  wonFight.Level.ChapterTitle,
			},
			Stars:       wonFight.Stars,
			DurationMS:  wonFight.DurationMS,
			DamageDealt: wonFight.DamageDealt,
			DamageTaken: wonFight.DamageTaken,
			FinishedAt:  wonFight.FinishedAt,
		})
	}
	response.WriteList(context, http.StatusOK, pagination.Page[WonFightResponse]{Items: wonFightResponses, Info: wonFightPage.Info})
}
