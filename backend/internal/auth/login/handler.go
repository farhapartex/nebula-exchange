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
	TwoFactorRequired    bool          `json:"two_factor_required"`
	AccessToken          string        `json:"access_token"`
	AccessTokenExpiresAt time.Time     `json:"access_token_expires_at"`
	User                 users.Profile `json:"user"`
}

type TwoFactorChallengeResponse struct {
	TwoFactorRequired  bool      `json:"two_factor_required"`
	ChallengeToken     string    `json:"challenge_token"`
	ChallengeExpiresAt time.Time `json:"challenge_expires_at"`
}

type Handler struct {
	service          *Service
	cookieSettings   session.CookieSettings
	now              func() time.Time
	loginRouteGuards []gin.HandlerFunc
}

func NewHandler(service *Service, cookieSettings session.CookieSettings, now func() time.Time, loginRouteGuards ...gin.HandlerFunc) *Handler {
	return &Handler{service: service, cookieSettings: cookieSettings, now: now, loginRouteGuards: loginRouteGuards}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	router.POST("/auth/login", append(handler.loginRouteGuards, handler.postLogin)...)
	router.POST("/auth/login/2fa", append(handler.loginRouteGuards, handler.postTwoFactorLogin)...)
	router.POST("/auth/refresh", handler.postRefresh)
	router.POST("/auth/logout", handler.postLogout)
}

type LogoutResponse struct {
	LoggedOut bool `json:"logged_out"`
}

func (handler *Handler) postLogout(context *gin.Context) {
	err := handler.service.LogOut(context.Request.Context(), session.RefreshTokenFromCookie(context))
	handler.cookieSettings.ClearRefreshCookie(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, LogoutResponse{LoggedOut: true})
}

func (handler *Handler) postLogin(context *gin.Context) {
	var loginRequest Request
	if err := request.BindJSON(context, &loginRequest); err != nil {
		response.WriteError(context, err)
		return
	}

	loginOutcome, err := handler.service.LogIn(context.Request.Context(), loginRequest, clientMetadataFrom(context))
	if err != nil {
		response.WriteError(context, err)
		return
	}
	if loginOutcome.TwoFactorChallenge != nil {
		response.WriteData(context, http.StatusOK, TwoFactorChallengeResponse{
			TwoFactorRequired:  true,
			ChallengeToken:     loginOutcome.TwoFactorChallenge.Token,
			ChallengeExpiresAt: loginOutcome.TwoFactorChallenge.ExpiresAt,
		})
		return
	}
	handler.respondWithSession(context, *loginOutcome.Session)
}

func (handler *Handler) postTwoFactorLogin(context *gin.Context) {
	var twoFactorRequest TwoFactorRequest
	if err := request.BindJSON(context, &twoFactorRequest); err != nil {
		response.WriteError(context, err)
		return
	}
	establishedSession, err := handler.service.CompleteTwoFactor(context.Request.Context(), twoFactorRequest, clientMetadataFrom(context))
	if err != nil {
		response.WriteError(context, err)
		return
	}
	handler.respondWithSession(context, establishedSession)
}

func (handler *Handler) postRefresh(context *gin.Context) {
	establishedSession, err := handler.service.Refresh(context.Request.Context(), session.RefreshTokenFromCookie(context), clientMetadataFrom(context))
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

func clientMetadataFrom(context *gin.Context) session.ClientMetadata {
	return session.ClientMetadata{UserAgent: context.GetHeader("User-Agent"), IPAddress: context.ClientIP()}
}
