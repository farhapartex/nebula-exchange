package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/service"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
)

const (
	maximumWebhookBodyBytes = 512 * 1024
	stripeSignatureHeader   = "Stripe-Signature"
)

type webhookReceipt struct {
	Received bool `json:"received"`
}

type StripeWebhookHandler struct {
	webhookService service.StripeWebhookService
}

func NewStripeWebhookHandler(webhookService service.StripeWebhookService) *StripeWebhookHandler {
	return &StripeWebhookHandler{webhookService: webhookService}
}

func (handler *StripeWebhookHandler) RegisterRoutes(router gin.IRouter) {
	router.POST("/webhooks/stripe", handler.postStripeEvent)
}

func (handler *StripeWebhookHandler) postStripeEvent(context *gin.Context) {
	payload, err := io.ReadAll(http.MaxBytesReader(context.Writer, context.Request.Body, maximumWebhookBodyBytes))
	if err != nil {
		response.WriteError(context, apierror.BadRequest("Request body could not be read"))
		return
	}
	if err := handler.webhookService.Handle(context.Request.Context(), payload, context.GetHeader(stripeSignatureHeader)); err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, webhookReceipt{Received: true})
}
