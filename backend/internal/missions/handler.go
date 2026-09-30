package missions

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/httpserver/request"
	"nebula-exchange/backend/internal/platform/httpserver/response"
)

type startMissionBody struct {
	ZoneID      string `json:"zone_id" binding:"required,max=64"`
	ShipItemID  int    `json:"ship_item_id" binding:"required,min=1"`
	DrillItemID int    `json:"drill_item_id" binding:"required,min=1"`
}

var knownStatuses = map[string]bool{
	string(StatusRunning): true, string(StatusCompleted): true, string(StatusCollected): true, string(StatusAborted): true,
}

type Handler struct {
	service      *Service
	accountGuard gin.HandlerFunc
}

func NewHandler(service *Service, accountGuard gin.HandlerFunc) *Handler {
	return &Handler{service: service, accountGuard: accountGuard}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	missionRoutes := router.Group("/missions", authentication.RequireUser())
	missionRoutes.POST("", handler.accountGuard, handler.startMission)
	missionRoutes.GET("", handler.listMissions)
	missionRoutes.GET("/:missionID", handler.getMission)
	missionRoutes.POST("/:missionID/collect", handler.accountGuard, handler.collectMission)
	missionRoutes.POST("/:missionID/abort", handler.accountGuard, handler.abortMission)
}

func (handler *Handler) startMission(context *gin.Context) {
	var body startMissionBody
	if err := request.BindJSON(context, &body); err != nil {
		response.WriteError(context, err)
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	mission, err := handler.service.Start(context.Request.Context(), userID, StartRequest{
		ZoneID: body.ZoneID, ShipItemID: body.ShipItemID, DrillItemID: body.DrillItemID,
	})
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusCreated, mission)
}

func (handler *Handler) listMissions(context *gin.Context) {
	pageRequest, err := request.PaginationFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	var statuses []string
	if statusFilter := context.Query("status"); statusFilter != "" {
		for _, status := range strings.Split(statusFilter, ",") {
			if !knownStatuses[status] {
				response.WriteError(context, apierror.ValidationFailed(map[string]string{"status": "must be RUNNING, COMPLETED, COLLECTED or ABORTED"}))
				return
			}
			statuses = append(statuses, status)
		}
	}
	var cursor *ListCursor
	if pageRequest.HasCursor() {
		decodedCursor, err := request.DecodeCursorPosition[ListCursor](pageRequest)
		if err != nil {
			response.WriteError(context, err)
			return
		}
		cursor = &decodedCursor
	}
	userID, _ := authentication.UserIDFrom(context)
	missionPage, err := handler.service.List(context.Request.Context(), userID, statuses, pageRequest, cursor)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteList(context, http.StatusOK, missionPage)
}

func (handler *Handler) getMission(context *gin.Context) {
	handler.withMissionID(context, func(userID, missionID uuid.UUID) (Mission, error) {
		return handler.service.Get(context.Request.Context(), userID, missionID)
	})
}

func (handler *Handler) collectMission(context *gin.Context) {
	handler.withMissionID(context, func(userID, missionID uuid.UUID) (Mission, error) {
		return handler.service.Collect(context.Request.Context(), userID, missionID)
	})
}

func (handler *Handler) abortMission(context *gin.Context) {
	handler.withMissionID(context, func(userID, missionID uuid.UUID) (Mission, error) {
		return handler.service.Abort(context.Request.Context(), userID, missionID)
	})
}

func (handler *Handler) withMissionID(context *gin.Context, action func(userID, missionID uuid.UUID) (Mission, error)) {
	missionID, err := uuid.Parse(context.Param("missionID"))
	if err != nil {
		response.WriteError(context, apierror.NotFound("This mission does not exist"))
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	mission, err := action(userID, missionID)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, mission)
}
