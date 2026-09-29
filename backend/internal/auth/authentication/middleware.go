package authentication

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/httpserver/response"
	"nebula-exchange/backend/internal/platform/idempotency"
)

const (
	userIDContextKey    = "authenticated_user_id"
	bearerPrefix        = "Bearer "
	authorizationHeader = "Authorization"
)

func IdentifyUser(tokenManager *accesstoken.Manager) gin.HandlerFunc {
	return func(context *gin.Context) {
		authorizationValue := context.GetHeader(authorizationHeader)
		if strings.HasPrefix(authorizationValue, bearerPrefix) {
			userID, err := tokenManager.Verify(strings.TrimPrefix(authorizationValue, bearerPrefix))
			if err == nil {
				context.Set(userIDContextKey, userID)
				idempotency.SetScope(context, "user:"+userID.String())
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
