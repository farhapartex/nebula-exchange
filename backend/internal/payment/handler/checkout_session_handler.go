package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment/service"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/authentication"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/request"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
)

type startCheckoutRequest struct {
	PlanID        string               `json:"plan_id" binding:"required,max=64"`
	ChapterCount  int                  `json:"chapter_count" binding:"required,min=1,max=1000"`
	PaymentMethod models.PaymentMethod `json:"payment_method" binding:"omitempty,oneof=CARD WALLET"`
}

type reportTransactionRequest struct {
	TransactionHash string `json:"transaction_hash" binding:"required,max=66"`
}

type CreatedWalletCheckoutResponse struct {
	ID               uuid.UUID `json:"id"`
	PaymentReference string    `json:"payment_reference"`
	USDCents         string    `json:"usd_cents"`
	Deadline         string    `json:"deadline"`
	Signature        string    `json:"signature"`
	VaultAddress     string    `json:"vault_address"`
	ChainID          int64     `json:"chain_id"`
	ExpiresAt        time.Time `json:"expires_at"`
}

type CreatedCheckoutSessionResponse struct {
	ID          uuid.UUID `json:"id"`
	CheckoutURL string    `json:"checkout_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type CheckoutSessionStateResponse struct {
	ID                    uuid.UUID              `json:"id"`
	Status                service.CheckoutStatus `json:"status"`
	Confirmations         *int                   `json:"confirmations"`
	RequiredConfirmations *int                   `json:"required_confirmations"`
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
	router.POST("/checkout-sessions/:checkout_session_id/transactions", authentication.RequireUser(), handler.postCheckoutTransaction)
}

func (handler *CheckoutSessionHandler) postCheckoutSession(context *gin.Context) {
	var body startCheckoutRequest
	if err := request.BindJSON(context, &body); err != nil {
		response.WriteError(context, err)
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	if body.PaymentMethod == models.PaymentMethodWallet {
		handler.startWalletCheckout(context, userID, body)
		return
	}
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
	response.WriteData(context, http.StatusOK, toCheckoutStateResponse(checkoutState))
}

func (handler *CheckoutSessionHandler) startWalletCheckout(context *gin.Context, userID uuid.UUID, body startCheckoutRequest) {
	startedCheckout, err := handler.checkoutService.StartWallet(context.Request.Context(), userID, body.PlanID, body.ChapterCount)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	statusCode := http.StatusCreated
	if startedCheckout.IsReused {
		statusCode = http.StatusOK
	}
	response.WriteData(context, statusCode, CreatedWalletCheckoutResponse{
		ID:               startedCheckout.ID,
		PaymentReference: startedCheckout.PaymentReference,
		USDCents:         formatCents(startedCheckout.USDCents),
		Deadline:         strconv.FormatInt(startedCheckout.Deadline.Unix(), 10),
		Signature:        startedCheckout.Signature,
		VaultAddress:     startedCheckout.VaultAddress,
		ChainID:          startedCheckout.ChainID,
		ExpiresAt:        startedCheckout.Deadline,
	})
}

func (handler *CheckoutSessionHandler) postCheckoutTransaction(context *gin.Context) {
	paymentID, err := uuid.Parse(context.Param("checkout_session_id"))
	if err != nil {
		response.WriteError(context, service.ErrCheckoutNotFound)
		return
	}
	var body reportTransactionRequest
	if err := request.BindJSON(context, &body); err != nil {
		response.WriteError(context, err)
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	checkoutState, err := handler.checkoutService.ReportWalletTransaction(context.Request.Context(), userID, paymentID, body.TransactionHash)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, toCheckoutStateResponse(checkoutState))
}

func toCheckoutStateResponse(checkoutState service.CheckoutState) CheckoutSessionStateResponse {
	return CheckoutSessionStateResponse{
		ID:                    checkoutState.ID,
		Status:                checkoutState.Status,
		Confirmations:         checkoutState.Confirmations,
		RequiredConfirmations: checkoutState.RequiredConfirmations,
	}
}
