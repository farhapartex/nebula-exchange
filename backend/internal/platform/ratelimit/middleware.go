package ratelimit

import (
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/httpserver/response"
)

var (
	SignupPolicy = Policy{Name: "signup", Limit: 5, Window: time.Hour}
	LoginPolicy  = Policy{Name: "login", Limit: 10, Window: 15 * time.Minute}
)

type MiddlewareFactory struct {
	limiter *SlidingWindowLimiter
	logger  *slog.Logger
}

func NewMiddlewareFactory(limiter *SlidingWindowLimiter, logger *slog.Logger) *MiddlewareFactory {
	return &MiddlewareFactory{limiter: limiter, logger: logger}
}

func (factory *MiddlewareFactory) PerClientIP(policy Policy) gin.HandlerFunc {
	return func(context *gin.Context) {
		decision, err := factory.limiter.Allow(context.Request.Context(), policy, context.ClientIP())
		if err != nil {
			factory.logger.ErrorContext(context.Request.Context(), "rate limiter unavailable, allowing request",
				slog.String("policy", policy.Name),
				slog.String("error", err.Error()),
			)
			context.Next()
			return
		}

		context.Header("X-RateLimit-Limit", strconv.Itoa(policy.Limit))
		context.Header("X-RateLimit-Remaining", strconv.Itoa(decision.Remaining))
		if decision.IsAllowed {
			context.Next()
			return
		}

		retryAfterSeconds := int(math.Ceil(decision.RetryAfter.Seconds()))
		context.Header("Retry-After", strconv.Itoa(retryAfterSeconds))
		response.WriteError(context, apierror.New(http.StatusTooManyRequests, apierror.CodeRateLimited,
			"Too many attempts. Please wait and try again.").WithDetails(map[string]int{"retry_after_seconds": retryAfterSeconds}))
	}
}
