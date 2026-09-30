package payments

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/httpserver/request"
	"nebula-exchange/backend/internal/platform/httpserver/response"
	"nebula-exchange/backend/internal/purpose"
)

type createPaymentBody struct {
	Purpose   string `json:"purpose" binding:"required,oneof=ENTRY_FEE TOPUP SHOP_PURCHASE UPGRADE_PURCHASE"`
	Method    string `json:"method" binding:"required,oneof=card crypto"`
	AmountNC  string `json:"amount_nc" binding:"required_if=Purpose TOPUP,max=20"`
	SKU       string `json:"sku" binding:"required_if=Purpose SHOP_PURCHASE,max=64"`
	UpgradeID string `json:"upgrade_id" binding:"required_if=Purpose UPGRADE_PURCHASE,max=64"`
	Quantity  int    `json:"quantity" binding:"omitempty,min=1,max=100"`
}

type Handler struct {
	service      *Service
	accountGuard gin.HandlerFunc
	rateLimit    gin.HandlerFunc
}

func NewHandler(service *Service, accountGuard gin.HandlerFunc, rateLimit gin.HandlerFunc) *Handler {
	return &Handler{service: service, accountGuard: accountGuard, rateLimit: rateLimit}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	paymentRoutes := router.Group("/payments", authentication.RequireUser())
	paymentRoutes.POST("", handler.rateLimit, handler.accountGuard, handler.createPayment)
	paymentRoutes.GET("", handler.listPayments)
	paymentRoutes.GET("/:paymentID", handler.getPayment)
}

func (handler *Handler) createPayment(context *gin.Context) {
	var body createPaymentBody
	if err := request.BindJSON(context, &body); err != nil {
		response.WriteError(context, err)
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	payment, err := handler.service.Create(context.Request.Context(), userID, CreateRequest{
		Purpose:   purpose.Kind(body.Purpose),
		Method:    Method(body.Method),
		AmountNC:  body.AmountNC,
		SKU:       body.SKU,
		UpgradeID: body.UpgradeID,
		Quantity:  body.Quantity,
	})
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusCreated, payment)
}

func (handler *Handler) getPayment(context *gin.Context) {
	paymentID, err := uuid.Parse(context.Param("paymentID"))
	if err != nil {
		response.WriteError(context, apierror.NotFound("This payment does not exist"))
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	payment, err := handler.service.Get(context.Request.Context(), userID, paymentID)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, payment)
}

func (handler *Handler) listPayments(context *gin.Context) {
	pageRequest, err := request.PaginationFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
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
	paymentPage, err := handler.service.List(context.Request.Context(), userID, pageRequest, cursor)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteList(context, http.StatusOK, paymentPage)
}
