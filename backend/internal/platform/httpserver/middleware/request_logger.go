package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(context *gin.Context) {
		requestStartedAt := time.Now()
		context.Next()

		statusCode := context.Writer.Status()
		logLevel := slog.LevelInfo
		if statusCode >= 500 {
			logLevel = slog.LevelError
		} else if statusCode >= 400 {
			logLevel = slog.LevelWarn
		}

		logger.LogAttrs(context.Request.Context(), logLevel, "http request",
			slog.String("request_id", RequestIDFrom(context)),
			slog.String("method", context.Request.Method),
			slog.String("path", context.FullPath()),
			slog.Int("status", statusCode),
			slog.Duration("duration", time.Since(requestStartedAt)),
			slog.String("client_ip", context.ClientIP()),
		)
	}
}
