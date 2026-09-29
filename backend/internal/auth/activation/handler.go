package activation

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/platform/httpserver/request"
	"nebula-exchange/backend/internal/platform/httpserver/response"
)

type activateRequest struct {
	Token string `json:"token" binding:"required,max=128"`
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	router.GET("/auth/activations/:token", handler.getActivation)
	router.POST("/auth/activations", handler.postActivation)
}

func (handler *Handler) getActivation(context *gin.Context) {
	activationPreview, err := handler.service.Preview(context.Request.Context(), context.Param("token"))
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, activationPreview)
}

func (handler *Handler) postActivation(context *gin.Context) {
	var activationRequest activateRequest
	if err := request.BindJSON(context, &activationRequest); err != nil {
		response.WriteError(context, err)
		return
	}
	activatedAccount, err := handler.service.Activate(context.Request.Context(), activationRequest.Token)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, activatedAccount)
}
