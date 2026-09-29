package balances

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/platform/httpserver/response"
)

type Handler struct {
	reader *Reader
}

func NewHandler(reader *Reader) *Handler {
	return &Handler{reader: reader}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	router.GET("/me/balances", authentication.RequireUser(), handler.getBalances)
}

func (handler *Handler) getBalances(context *gin.Context) {
	userID, _ := authentication.UserIDFrom(context)
	summary, err := handler.reader.Summarize(context.Request.Context(), userID)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, summary)
}
