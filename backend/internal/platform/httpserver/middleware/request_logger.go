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

		routePath := context.FullPath()
		if routePath == "" {
			routePath = context.Request.URL.Path
		}

		requestAttributes := []slog.Attr{
			slog.String("request_id", RequestIDFrom(context)),
			slog.String("method", context.Request.Method),
			slog.String("path", routePath),
			slog.Int("status", statusCode),
			slog.Duration("duration", time.Since(requestStartedAt)),
			slog.String("client_ip", context.ClientIP()),
		}
		if len(context.Errors) > 0 {
			requestAttributes = append(requestAttributes, slog.String("error", context.Errors.String()))
		}
		logger.LogAttrs(context.Request.Context(), logLevel, "http request", requestAttributes...)
	}
}
