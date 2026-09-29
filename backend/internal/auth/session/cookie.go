package session

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RefreshCookieName = "nebula_refresh_token"
	refreshCookiePath = "/api/v1/auth"
)

type CookieSettings struct {
	IsSecure bool
}

func (settings CookieSettings) SetRefreshCookie(context *gin.Context, issuedToken IssuedRefreshToken, now time.Time) {
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    issuedToken.Plaintext,
		Path:     refreshCookiePath,
		Expires:  issuedToken.ExpiresAt,
		MaxAge:   int(issuedToken.ExpiresAt.Sub(now).Seconds()),
		HttpOnly: true,
		Secure:   settings.IsSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (settings CookieSettings) ClearRefreshCookie(context *gin.Context) {
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     refreshCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   settings.IsSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func RefreshTokenFromCookie(context *gin.Context) string {
	cookieValue, err := context.Cookie(RefreshCookieName)
	if err != nil {
		return ""
	}
	return cookieValue
}
