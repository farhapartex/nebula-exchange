package passwordreset

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/platform/httpserver/request"
	"nebula-exchange/backend/internal/platform/httpserver/response"
	"nebula-exchange/backend/internal/platform/ratelimit"
)

type resetRequest struct {
	Email string `json:"email" binding:"required,max=254"`
}

type resetSubmission struct {
	Token    string `json:"token" binding:"required,max=128"`
	Password string `json:"password" binding:"required,min=10,max=128"`
}

type AcceptedResult struct {
	Accepted bool `json:"accepted"`
}

type ResetResult struct {
	PasswordReset bool `json:"password_reset"`
}

type Handler struct {
	service    *Service
	rateLimits *ratelimit.MiddlewareFactory
}

func NewHandler(service *Service, rateLimits *ratelimit.MiddlewareFactory) *Handler {
	return &Handler{service: service, rateLimits: rateLimits}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	router.POST("/auth/password-reset-requests", handler.rateLimits.PerClientIP(ratelimit.PasswordResetRequestPolicy), handler.postResetRequest)
	router.GET("/auth/password-resets/:token", handler.getReset)
	router.POST("/auth/password-resets", handler.rateLimits.PerClientIP(ratelimit.PasswordResetSubmissionPolicy), handler.postReset)
}

func (handler *Handler) postResetRequest(context *gin.Context) {
	var passwordResetRequest resetRequest
	if err := request.BindJSON(context, &passwordResetRequest); err != nil {
		response.WriteError(context, err)
		return
	}
	emailSubject := ratelimit.HashedSubject(normalizeEmail(passwordResetRequest.Email))
	if !handler.rateLimits.AllowOrReject(context, ratelimit.PasswordResetRequestEmailPolicy, emailSubject) {
		return
	}
	if err := handler.service.RequestReset(context.Request.Context(), passwordResetRequest.Email); err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusAccepted, AcceptedResult{Accepted: true})
}

func (handler *Handler) getReset(context *gin.Context) {
	resetPreview, err := handler.service.Preview(context.Request.Context(), context.Param("token"))
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, resetPreview)
}

func (handler *Handler) postReset(context *gin.Context) {
	var passwordResetSubmission resetSubmission
	if err := request.BindJSON(context, &passwordResetSubmission); err != nil {
		response.WriteError(context, err)
		return
	}
	if err := handler.service.ResetPassword(context.Request.Context(), passwordResetSubmission.Token, passwordResetSubmission.Password); err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, ResetResult{PasswordReset: true})
}
