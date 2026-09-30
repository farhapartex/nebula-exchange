package inventory

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/platform/httpserver/request"
	"nebula-exchange/backend/internal/platform/httpserver/response"
)

type Handler struct {
	reader *Reader
}

func NewHandler(reader *Reader) *Handler {
	return &Handler{reader: reader}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	router.GET("/me/inventory", authentication.RequireUser(), handler.listInventory)
}

func (handler *Handler) listInventory(context *gin.Context) {
	pageRequest, err := request.PaginationFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	var cursor Cursor
	if pageRequest.HasCursor() {
		if cursor, err = request.DecodeCursorPosition[Cursor](pageRequest); err != nil {
			response.WriteError(context, err)
			return
		}
	}
	userID, _ := authentication.UserIDFrom(context)
	holdingPage, err := handler.reader.List(context.Request.Context(), userID, pageRequest, cursor)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteList(context, http.StatusOK, holdingPage)
}
