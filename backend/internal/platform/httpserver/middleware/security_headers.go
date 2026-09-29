package middleware

import "github.com/gin-gonic/gin"

func SecurityHeaders(isProduction bool) gin.HandlerFunc {
	return func(context *gin.Context) {
		responseHeaders := context.Writer.Header()
		responseHeaders.Set("X-Content-Type-Options", "nosniff")
		responseHeaders.Set("X-Frame-Options", "DENY")
		responseHeaders.Set("Referrer-Policy", "no-referrer")
		responseHeaders.Set("Cache-Control", "no-store")
		if isProduction {
			responseHeaders.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		}
		context.Next()
	}
}
