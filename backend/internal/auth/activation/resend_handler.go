package activation

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/platform/httpserver/request"
	"nebula-exchange/backend/internal/platform/httpserver/response"
	"nebula-exchange/backend/internal/platform/ratelimit"
)

type resendRequest struct {
	Email string `json:"email" binding:"required,max=254"`
}

type ResendResult struct {
	Accepted bool `json:"accepted"`
}

type ResendHandler struct {
	resender   *Resender
	rateLimits *ratelimit.MiddlewareFactory
}

func NewResendHandler(resender *Resender, rateLimits *ratelimit.MiddlewareFactory) *ResendHandler {
	return &ResendHandler{resender: resender, rateLimits: rateLimits}
}

func (handler *ResendHandler) RegisterRoutes(router gin.IRouter) {
	router.POST("/auth/activation-emails", handler.rateLimits.PerClientIP(ratelimit.ResendActivationPolicy), handler.postActivationEmail)
}

func (handler *ResendHandler) postActivationEmail(context *gin.Context) {
	var resendActivationRequest resendRequest
	if err := request.BindJSON(context, &resendActivationRequest); err != nil {
		response.WriteError(context, err)
		return
	}

	emailSubject := ratelimit.HashedSubject(NormalizeEmail(resendActivationRequest.Email))
	if !handler.rateLimits.AllowOrReject(context, ratelimit.ResendActivationEmailPolicy, emailSubject) {
		return
	}

	if err := handler.resender.ResendActivationEmail(context.Request.Context(), resendActivationRequest.Email); err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusAccepted, ResendResult{Accepted: true})
}
