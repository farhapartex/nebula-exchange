package ledgerhistory

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/platform/apierror"
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
	router.GET("/me/ledger", authentication.RequireUser(), handler.listJournals)
}

func (handler *Handler) listJournals(context *gin.Context) {
	pageRequest, err := request.PaginationFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	filter := Filter{JournalType: context.Query("type")}
	if filter.JournalType != "" && !ledger.IsKnownJournalType(filter.JournalType) {
		response.WriteError(context, apierror.ValidationFailed(map[string]string{"type": "is not a known journal type"}))
		return
	}
	var cursor *Cursor
	if pageRequest.HasCursor() {
		decodedCursor, err := request.DecodeCursorPosition[Cursor](pageRequest)
		if err != nil {
			response.WriteError(context, err)
			return
		}
		cursor = &decodedCursor
	}
	userID, _ := authentication.UserIDFrom(context)
	journalPage, err := handler.reader.List(context.Request.Context(), userID, filter, pageRequest, cursor)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteList(context, http.StatusOK, journalPage)
}
