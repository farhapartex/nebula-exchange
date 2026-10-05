package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity/service"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/request"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
)

type activationRequest struct {
	Token string `json:"token" binding:"required,max=128"`
}

type ActivationPreviewResponse struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
	IsActivated bool   `json:"is_activated"`
}

type ActivatedAccountResponse struct {
	Email       string            `json:"email"`
	Username    string            `json:"username"`
	Status      models.UserStatus `json:"status"`
	ActivatedAt *time.Time        `json:"activated_at"`
}

type ActivationHandler struct {
	activationService service.ActivationService
}

func NewActivationHandler(activationService service.ActivationService) *ActivationHandler {
	return &ActivationHandler{activationService: activationService}
}

func (handler *ActivationHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/auth/activations/:token", handler.getActivation)
	router.POST("/auth/activations", handler.postActivation)
}

func (handler *ActivationHandler) getActivation(context *gin.Context) {
	activationPreview, err := handler.activationService.Preview(context.Request.Context(), context.Param("token"))
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, ActivationPreviewResponse{
		Email:       activationPreview.Email,
		Username:    activationPreview.Username,
		IsActivated: activationPreview.IsActivated,
	})
}

func (handler *ActivationHandler) postActivation(context *gin.Context) {
	var body activationRequest
	if err := request.BindJSON(context, &body); err != nil {
		response.WriteError(context, err)
		return
	}
	activatedUser, err := handler.activationService.Activate(context.Request.Context(), body.Token)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, ActivatedAccountResponse{
		Email:       activatedUser.Email,
		Username:    activatedUser.Username,
		Status:      activatedUser.Status,
		ActivatedAt: activatedUser.ActivatedAt,
	})
}
