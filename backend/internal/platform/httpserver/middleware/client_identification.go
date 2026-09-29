package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/httpserver/response"
)

const (
	ClientIdentificationHeader = "X-Nebula-Client"
	WebClientIdentifier        = "web"
)

func RequireClientIdentification(exemptPathPrefixes ...string) gin.HandlerFunc {
	return func(context *gin.Context) {
		if !isStateChangingRequest(context.Request.Method) || hasExemptPrefix(context.Request.URL.Path, exemptPathPrefixes) {
			context.Next()
			return
		}
		if context.GetHeader(ClientIdentificationHeader) != WebClientIdentifier {
			response.WriteError(context, apierror.Forbidden("Missing or invalid "+ClientIdentificationHeader+" header"))
			return
		}
		context.Next()
	}
}

func isStateChangingRequest(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func hasExemptPrefix(path string, exemptPathPrefixes []string) bool {
	for _, exemptPathPrefix := range exemptPathPrefixes {
		if strings.HasPrefix(path, exemptPathPrefix) {
			return true
		}
	}
	return false
}
