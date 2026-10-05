package idempotency

import (
	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/authentication"
)

const anonymousScope = "anonymous"

func scopeFrom(context *gin.Context) string {
	if userID, isAuthenticated := authentication.UserIDFrom(context); isAuthenticated {
		return "user:" + userID.String()
	}
	return anonymousScope
}
