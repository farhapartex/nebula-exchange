package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/platform/httpserver/response"
)

func PanicRecovery(logger *slog.Logger) gin.HandlerFunc {
	return func(context *gin.Context) {
		defer func() {
			recoveredValue := recover()
			if recoveredValue == nil {
				return
			}
			logger.ErrorContext(context.Request.Context(), "panic recovered",
				slog.String("request_id", RequestIDFrom(context)),
				slog.Any("panic", recoveredValue),
				slog.String("stack", string(debug.Stack())),
			)
			response.AbortWithError(context, http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong")
		}()
		context.Next()
	}
}
