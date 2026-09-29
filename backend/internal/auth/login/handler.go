package login

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/auth/session"
	"nebula-exchange/backend/internal/platform/httpserver/request"
	"nebula-exchange/backend/internal/platform/httpserver/response"
	"nebula-exchange/backend/internal/users"
)

type SessionResponse struct {
	AccessToken          string        `json:"access_token"`
	AccessTokenExpiresAt time.Time     `json:"access_token_expires_at"`
	User                 users.Profile `json:"user"`
}

type Handler struct {
	service        *Service
	cookieSettings session.CookieSettings
	now            func() time.Time
}

func NewHandler(service *Service, cookieSettings session.CookieSettings, now func() time.Time) *Handler {
	return &Handler{service: service, cookieSettings: cookieSettings, now: now}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	router.POST("/auth/login", handler.postLogin)
	router.POST("/auth/refresh", handler.postRefresh)
}

func (handler *Handler) postLogin(context *gin.Context) {
	var loginRequest Request
	if err := request.BindJSON(context, &loginRequest); err != nil {
		response.WriteError(context, err)
		return
	}

	establishedSession, err := handler.service.LogIn(context.Request.Context(), loginRequest)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	handler.respondWithSession(context, establishedSession)
}

func (handler *Handler) postRefresh(context *gin.Context) {
	establishedSession, err := handler.service.Refresh(context.Request.Context(), session.RefreshTokenFromCookie(context))
	if err != nil {
		handler.cookieSettings.ClearRefreshCookie(context)
		response.WriteError(context, err)
		return
	}
	handler.respondWithSession(context, establishedSession)
}

func (handler *Handler) respondWithSession(context *gin.Context, establishedSession EstablishedSession) {
	handler.cookieSettings.SetRefreshCookie(context, establishedSession.RefreshToken, handler.now())
	response.WriteData(context, http.StatusOK, SessionResponse{
		AccessToken:          establishedSession.AccessToken.Value,
		AccessTokenExpiresAt: establishedSession.AccessToken.ExpiresAt,
		User:                 establishedSession.User.Profile(),
	})
}
