package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity/service"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/request"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
)

var NonReplayableRoutes = []string{"/auth/login", "/auth/refresh", "/auth/logout"}

type loginRequest struct {
	Email    string `json:"email" binding:"required,max=254"`
	Password string `json:"password" binding:"required,max=128"`
}

type SessionResponse struct {
	AccessToken          string              `json:"access_token"`
	AccessTokenExpiresAt time.Time           `json:"access_token_expires_at"`
	User                 UserProfileResponse `json:"user"`
}

type LogoutResponse struct {
	LoggedOut bool `json:"logged_out"`
}

type SessionHandler struct {
	sessionService service.SessionService
	refreshCookie  RefreshCookie
}

func NewSessionHandler(sessionService service.SessionService, refreshCookie RefreshCookie) *SessionHandler {
	return &SessionHandler{sessionService: sessionService, refreshCookie: refreshCookie}
}

func (handler *SessionHandler) RegisterRoutes(router gin.IRouter) {
	router.POST("/auth/login", handler.postLogin)
	router.POST("/auth/refresh", handler.postRefresh)
	router.POST("/auth/logout", handler.postLogout)
}

func (handler *SessionHandler) postLogin(context *gin.Context) {
	var body loginRequest
	if err := request.BindJSON(context, &body); err != nil {
		response.WriteError(context, err)
		return
	}
	establishedSession, err := handler.sessionService.LogIn(context.Request.Context(), service.LoginInput{
		Email:    body.Email,
		Password: body.Password,
	}, clientMetadataFrom(context))
	if err != nil {
		response.WriteError(context, err)
		return
	}
	handler.respondWithSession(context, establishedSession)
}

func (handler *SessionHandler) postRefresh(context *gin.Context) {
	establishedSession, err := handler.sessionService.Refresh(context.Request.Context(), handler.refreshCookie.Read(context), clientMetadataFrom(context))
	if err != nil {
		handler.refreshCookie.Clear(context)
		response.WriteError(context, err)
		return
	}
	handler.respondWithSession(context, establishedSession)
}

func (handler *SessionHandler) postLogout(context *gin.Context) {
	err := handler.sessionService.LogOut(context.Request.Context(), handler.refreshCookie.Read(context))
	handler.refreshCookie.Clear(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, LogoutResponse{LoggedOut: true})
}

func (handler *SessionHandler) respondWithSession(context *gin.Context, establishedSession service.EstablishedSession) {
	handler.refreshCookie.Set(context, establishedSession.RefreshToken)
	response.WriteData(context, http.StatusOK, SessionResponse{
		AccessToken:          establishedSession.AccessToken.Value,
		AccessTokenExpiresAt: establishedSession.AccessToken.ExpiresAt,
		User:                 toUserProfileResponse(establishedSession.User),
	})
}

func clientMetadataFrom(context *gin.Context) service.ClientMetadata {
	return service.ClientMetadata{UserAgent: context.GetHeader("User-Agent"), IPAddress: context.ClientIP()}
}
