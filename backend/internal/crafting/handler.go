package crafting

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/httpserver/request"
	"nebula-exchange/backend/internal/platform/httpserver/response"
)

type startCraftBody struct {
	RecipeID string `json:"recipe_id" binding:"required,max=64"`
	Quantity int    `json:"quantity" binding:"required,min=1,max=100"`
}

type Handler struct {
	service      *Service
	accountGuard gin.HandlerFunc
}

func NewHandler(service *Service, accountGuard gin.HandlerFunc) *Handler {
	return &Handler{service: service, accountGuard: accountGuard}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	craftRoutes := router.Group("/crafts", authentication.RequireUser())
	craftRoutes.POST("", handler.accountGuard, handler.startCraft)
	craftRoutes.GET("", handler.listCrafts)
}

func (handler *Handler) startCraft(context *gin.Context) {
	var body startCraftBody
	if err := request.BindJSON(context, &body); err != nil {
		response.WriteError(context, err)
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	craftJob, err := handler.service.Start(context.Request.Context(), userID, body.RecipeID, body.Quantity)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusCreated, craftJob)
}

func (handler *Handler) listCrafts(context *gin.Context) {
	pageRequest, err := request.PaginationFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	status := context.Query("status")
	if status != "" && status != string(StatusCrafting) && status != string(StatusDelivered) {
		response.WriteError(context, apierror.ValidationFailed(map[string]string{"status": "must be CRAFTING or DELIVERED"}))
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
	craftPage, err := handler.service.List(context.Request.Context(), userID, status, pageRequest, cursor)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteList(context, http.StatusOK, craftPage)
}
