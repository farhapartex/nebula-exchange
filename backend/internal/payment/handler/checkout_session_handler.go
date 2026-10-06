package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/service"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/authentication"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/request"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
)

type startCheckoutRequest struct {
	PlanID       string `json:"plan_id" binding:"required,max=64"`
	ChapterCount int    `json:"chapter_count" binding:"required,min=1,max=1000"`
}

type CreatedCheckoutSessionResponse struct {
	ID          uuid.UUID `json:"id"`
	CheckoutURL string    `json:"checkout_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type CheckoutSessionStateResponse struct {
	ID     uuid.UUID              `json:"id"`
	Status service.CheckoutStatus `json:"status"`
}

type CheckoutSessionHandler struct {
	checkoutService service.CheckoutService
}

func NewCheckoutSessionHandler(checkoutService service.CheckoutService) *CheckoutSessionHandler {
	return &CheckoutSessionHandler{checkoutService: checkoutService}
}

func (handler *CheckoutSessionHandler) RegisterRoutes(router gin.IRouter) {
	router.POST("/checkout-sessions", authentication.RequireUser(), handler.postCheckoutSession)
	router.GET("/checkout-sessions/:checkout_session_id", authentication.RequireUser(), handler.getCheckoutSession)
}

func (handler *CheckoutSessionHandler) postCheckoutSession(context *gin.Context) {
	var body startCheckoutRequest
	if err := request.BindJSON(context, &body); err != nil {
		response.WriteError(context, err)
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	startedCheckout, err := handler.checkoutService.Start(context.Request.Context(), userID, body.PlanID, body.ChapterCount)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusCreated, CreatedCheckoutSessionResponse{
		ID:          startedCheckout.ID,
		CheckoutURL: startedCheckout.CheckoutURL,
		ExpiresAt:   startedCheckout.ExpiresAt,
	})
}

func (handler *CheckoutSessionHandler) getCheckoutSession(context *gin.Context) {
	paymentID, err := uuid.Parse(context.Param("checkout_session_id"))
	if err != nil {
		response.WriteError(context, service.ErrCheckoutNotFound)
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	checkoutState, err := handler.checkoutService.State(context.Request.Context(), userID, paymentID)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, CheckoutSessionStateResponse{ID: checkoutState.ID, Status: checkoutState.Status})
}
