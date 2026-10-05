package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity/service"
)

const (
	refreshCookieName = "street_born_refresh_token"
	refreshCookiePath = "/api/v1/auth"
)

type RefreshCookie struct {
	IsSecure bool
	Now      func() time.Time
}

func (cookie RefreshCookie) Set(context *gin.Context, refreshToken service.IssuedRefreshToken) {
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken.Plaintext,
		Path:     refreshCookiePath,
		Expires:  refreshToken.ExpiresAt,
		MaxAge:   int(refreshToken.ExpiresAt.Sub(cookie.Now()).Seconds()),
		HttpOnly: true,
		Secure:   cookie.IsSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (cookie RefreshCookie) Clear(context *gin.Context) {
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     refreshCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cookie.IsSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (cookie RefreshCookie) Read(context *gin.Context) string {
	cookieValue, err := context.Cookie(refreshCookieName)
	if err != nil {
		return ""
	}
	return cookieValue
}
