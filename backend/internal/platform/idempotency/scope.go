package idempotency

import "github.com/gin-gonic/gin"

const (
	scopeContextKey = "idempotency_scope"
	anonymousScope  = "anonymous"
)

func SetScope(context *gin.Context, scope string) {
	context.Set(scopeContextKey, scope)
}

func scopeFrom(context *gin.Context) string {
	if scope := context.GetString(scopeContextKey); scope != "" {
		return scope
	}
	return anonymousScope
}
