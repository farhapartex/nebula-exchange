package session

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/httpserver/request"
	"nebula-exchange/backend/internal/platform/httpserver/response"
)

type RevokeResult struct {
	Revoked           bool `json:"revoked"`
	WasCurrentSession bool `json:"was_current_session"`
}

type SessionsHandler struct {
	sessions       *Sessions
	cookieSettings CookieSettings
}

func NewSessionsHandler(sessions *Sessions, cookieSettings CookieSettings) *SessionsHandler {
	return &SessionsHandler{sessions: sessions, cookieSettings: cookieSettings}
}

func (handler *SessionsHandler) RegisterRoutes(router gin.IRouter) {
	sessionRoutes := router.Group("/auth/sessions", authentication.RequireUser())
	sessionRoutes.GET("", handler.listSessions)
	sessionRoutes.DELETE("/:sessionID", handler.revokeSession)
}

func (handler *SessionsHandler) listSessions(context *gin.Context) {
	pageRequest, err := request.PaginationFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	var cursorPosition *sessionCursor
	if pageRequest.HasCursor() {
		decodedCursor, err := request.DecodeCursorPosition[sessionCursor](pageRequest)
		if err != nil {
			response.WriteError(context, err)
			return
		}
		cursorPosition = &decodedCursor
	}

	userID, _ := authentication.UserIDFrom(context)
	sessionPage, err := handler.sessions.List(context.Request.Context(), userID, currentTokenHash(context), pageRequest, cursorPosition)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteList(context, http.StatusOK, sessionPage)
}

func (handler *SessionsHandler) revokeSession(context *gin.Context) {
	sessionID, err := uuid.Parse(context.Param("sessionID"))
	if err != nil {
		response.WriteError(context, apierror.NotFound("This session no longer exists"))
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	revocation, err := handler.sessions.Revoke(context.Request.Context(), userID, sessionID, currentTokenHash(context))
	if err != nil {
		response.WriteError(context, err)
		return
	}
	if revocation.WasCurrentSession {
		handler.cookieSettings.ClearRefreshCookie(context)
	}
	response.WriteData(context, http.StatusOK, RevokeResult{Revoked: true, WasCurrentSession: revocation.WasCurrentSession})
}

func currentTokenHash(context *gin.Context) []byte {
	return CurrentTokenHash(context)
}

func CurrentTokenHash(context *gin.Context) []byte {
	plaintextToken := RefreshTokenFromCookie(context)
	if plaintextToken == "" {
		return nil
	}
	return HashForCookie(plaintextToken)
}
