package twofactor

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/platform/httpserver/request"
	"nebula-exchange/backend/internal/platform/httpserver/response"
)

type codeSubmission struct {
	Code string `json:"code" binding:"required,len=6,numeric"`
}

type ToggleResult struct {
	TwoFactorEnabled bool `json:"two_factor_enabled"`
}

type Handler struct {
	service     *Service
	routeGuards []gin.HandlerFunc
}

func NewHandler(service *Service, routeGuards ...gin.HandlerFunc) *Handler {
	return &Handler{service: service, routeGuards: routeGuards}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	protectedRoutes := router.Group("/auth/2fa", append([]gin.HandlerFunc{authentication.RequireUser()}, handler.routeGuards...)...)
	protectedRoutes.POST("/setup", handler.postSetup)
	protectedRoutes.POST("/enable", handler.postEnable)
	protectedRoutes.POST("/disable", handler.postDisable)
}

func (handler *Handler) postSetup(context *gin.Context) {
	userID, _ := authentication.UserIDFrom(context)
	setupDetails, err := handler.service.BeginSetup(context.Request.Context(), userID)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, setupDetails)
}

func (handler *Handler) postEnable(context *gin.Context) {
	handler.withCode(context, func(code string) error {
		userID, _ := authentication.UserIDFrom(context)
		return handler.service.Enable(context.Request.Context(), userID, code)
	}, true)
}

func (handler *Handler) postDisable(context *gin.Context) {
	handler.withCode(context, func(code string) error {
		userID, _ := authentication.UserIDFrom(context)
		return handler.service.Disable(context.Request.Context(), userID, code)
	}, false)
}

func (handler *Handler) withCode(context *gin.Context, applyCode func(code string) error, isEnabledAfter bool) {
	var submission codeSubmission
	if err := request.BindJSON(context, &submission); err != nil {
		response.WriteError(context, err)
		return
	}
	if err := applyCode(submission.Code); err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, ToggleResult{TwoFactorEnabled: isEnabledAfter})
}
