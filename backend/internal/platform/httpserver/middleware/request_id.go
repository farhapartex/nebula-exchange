package middleware

import (
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	RequestIDHeader     = "X-Request-ID"
	requestIDContextKey = "request_id"
)

var acceptedRequestIDPattern = regexp.MustCompile(`^[A-Za-z0-9\-]{8,64}$`)

func RequestID() gin.HandlerFunc {
	return func(context *gin.Context) {
		requestID := context.GetHeader(RequestIDHeader)
		if !acceptedRequestIDPattern.MatchString(requestID) {
			requestID = uuid.NewString()
		}
		context.Set(requestIDContextKey, requestID)
		context.Header(RequestIDHeader, requestID)
		context.Next()
	}
}

func RequestIDFrom(context *gin.Context) string {
	return context.GetString(requestIDContextKey)
}
