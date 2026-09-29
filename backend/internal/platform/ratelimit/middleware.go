package ratelimit

import (
	"crypto/sha256"
	"encoding/hex"
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
	SignupPolicy                = Policy{Name: "signup", Limit: 5, Window: time.Hour}
	LoginPolicy                 = Policy{Name: "login", Limit: 10, Window: 15 * time.Minute}
	ResendActivationPolicy      = Policy{Name: "resend_activation", Limit: 3, Window: 15 * time.Minute}
	ResendActivationEmailPolicy = Policy{Name: "resend_activation_email", Limit: 1, Window: time.Minute}

	PasswordResetRequestPolicy      = Policy{Name: "password_reset_request", Limit: 3, Window: 15 * time.Minute}
	PasswordResetRequestEmailPolicy = Policy{Name: "password_reset_request_email", Limit: 1, Window: time.Minute}
	PasswordResetSubmissionPolicy   = Policy{Name: "password_reset_submission", Limit: 10, Window: 15 * time.Minute}

	TwoFactorChangePolicy = Policy{Name: "two_factor_change", Limit: 10, Window: 15 * time.Minute}
	PasswordChangePolicy  = Policy{Name: "password_change", Limit: 5, Window: 15 * time.Minute}
)

func HashedSubject(subjectValue string) string {
	subjectHash := sha256.Sum256([]byte(subjectValue))
	return hex.EncodeToString(subjectHash[:])
}

type MiddlewareFactory struct {
	limiter *SlidingWindowLimiter
	logger  *slog.Logger
}

func NewMiddlewareFactory(limiter *SlidingWindowLimiter, logger *slog.Logger) *MiddlewareFactory {
	return &MiddlewareFactory{limiter: limiter, logger: logger}
}

func (factory *MiddlewareFactory) PerClientIP(policy Policy) gin.HandlerFunc {
	return func(context *gin.Context) {
		if factory.AllowOrReject(context, policy, context.ClientIP()) {
			context.Next()
		}
	}
}

func (factory *MiddlewareFactory) AllowOrReject(context *gin.Context, policy Policy, subject string) bool {
	decision, err := factory.limiter.Allow(context.Request.Context(), policy, subject)
	if err != nil {
		factory.logger.ErrorContext(context.Request.Context(), "rate limiter unavailable, allowing request",
			slog.String("policy", policy.Name),
			slog.String("error", err.Error()),
		)
		return true
	}

	context.Header("X-RateLimit-Limit", strconv.Itoa(policy.Limit))
	context.Header("X-RateLimit-Remaining", strconv.Itoa(decision.Remaining))
	if decision.IsAllowed {
		return true
	}

	retryAfterSeconds := int(math.Ceil(decision.RetryAfter.Seconds()))
	context.Header("Retry-After", strconv.Itoa(retryAfterSeconds))
	response.WriteError(context, apierror.New(http.StatusTooManyRequests, apierror.CodeRateLimited,
		"Too many attempts. Please wait and try again.").WithDetails(map[string]int{"retry_after_seconds": retryAfterSeconds}))
	return false
}
