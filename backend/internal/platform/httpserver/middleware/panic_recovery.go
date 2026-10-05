package middleware

import (
	"log/slog"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
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
			response.WriteError(context, apierror.Internal())
		}()
		context.Next()
	}
}
