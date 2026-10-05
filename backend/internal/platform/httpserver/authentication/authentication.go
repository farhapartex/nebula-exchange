package authentication

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
)

const (
	userIDContextKey    = "authenticated_user_id"
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
)

type AccessTokenVerifier interface {
	Verify(accessToken string) (uuid.UUID, error)
}

func IdentifyUser(verifier AccessTokenVerifier) gin.HandlerFunc {
	return func(context *gin.Context) {
		authorizationValue := context.GetHeader(authorizationHeader)
		if strings.HasPrefix(authorizationValue, bearerPrefix) {
			userID, err := verifier.Verify(strings.TrimPrefix(authorizationValue, bearerPrefix))
			if err == nil {
				context.Set(userIDContextKey, userID)
			}
		}
		context.Next()
	}
}

func RequireUser() gin.HandlerFunc {
	return func(context *gin.Context) {
		if _, isAuthenticated := UserIDFrom(context); !isAuthenticated {
			response.WriteError(context, apierror.Unauthorized("Log in to continue"))
			return
		}
		context.Next()
	}
}

func UserIDFrom(context *gin.Context) (uuid.UUID, bool) {
	storedValue, isSet := context.Get(userIDContextKey)
	if !isSet {
		return uuid.Nil, false
	}
	userID, isUserID := storedValue.(uuid.UUID)
	return userID, isUserID
}
