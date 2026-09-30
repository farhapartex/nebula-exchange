package shop

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/platform/httpserver/request"
	"nebula-exchange/backend/internal/platform/httpserver/response"
)

type purchaseBody struct {
	Quantity int `json:"quantity" binding:"required,min=1,max=100"`
}

type Handler struct {
	service      *Service
	accountGuard gin.HandlerFunc
}

func NewHandler(service *Service, accountGuard gin.HandlerFunc) *Handler {
	return &Handler{service: service, accountGuard: accountGuard}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	router.POST("/shop-items/:sku/purchases", authentication.RequireUser(), handler.accountGuard, handler.purchase)
}

func (handler *Handler) purchase(context *gin.Context) {
	var body purchaseBody
	if err := request.BindJSON(context, &body); err != nil {
		response.WriteError(context, err)
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	completedPurchase, err := handler.service.BuyWithBalance(context.Request.Context(), userID, context.Param("sku"), body.Quantity)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusCreated, completedPurchase)
}
