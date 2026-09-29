package loginlockout

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"nebula-exchange/backend/internal/platform/apierror"
)

type Policy struct {
	MaximumFailures int
	FailureWindow   time.Duration
	LockDuration    time.Duration
}

var DefaultPolicy = Policy{MaximumFailures: 5, FailureWindow: 15 * time.Minute, LockDuration: 15 * time.Minute}

var recordFailureScript = redis.NewScript(`
local failureCount = redis.call("INCR", KEYS[1])
if failureCount == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
if failureCount >= tonumber(ARGV[2]) then
  redis.call("SET", KEYS[2], "1", "PX", ARGV[3])
  redis.call("DEL", KEYS[1])
  return tonumber(ARGV[3])
end
return 0
`)

type Guard struct {
	client    redis.Cmdable
	keyPrefix string
	policy    Policy
	logger    *slog.Logger
}

func NewGuard(client redis.Cmdable, keyPrefix string, policy Policy, logger *slog.Logger) *Guard {
	return &Guard{client: client, keyPrefix: keyPrefix, policy: policy, logger: logger}
}

func LockedError(retryAfter time.Duration) error {
	retryAfterSeconds := int(math.Ceil(retryAfter.Seconds()))
	return apierror.New(http.StatusTooManyRequests, apierror.CodeLoginLocked,
		"Too many failed login attempts. Try again later.").WithDetails(map[string]int{"retry_after_seconds": retryAfterSeconds})
}

func (guard *Guard) EnsureNotLocked(ctx context.Context, emailAddress string) error {
	remainingLock, err := guard.client.PTTL(ctx, guard.lockKey(emailAddress)).Result()
	if err != nil {
		guard.logUnavailable(ctx, err)
		return nil
	}
	if remainingLock > 0 {
		return LockedError(remainingLock)
	}
	return nil
}

func (guard *Guard) RecordFailure(ctx context.Context, emailAddress string) error {
	lockMilliseconds, err := recordFailureScript.Run(ctx, guard.client,
		[]string{guard.failureKey(emailAddress), guard.lockKey(emailAddress)},
		guard.policy.FailureWindow.Milliseconds(), guard.policy.MaximumFailures, guard.policy.LockDuration.Milliseconds(),
	).Int64()
	if err != nil {
		guard.logUnavailable(ctx, err)
		return nil
	}
	if lockMilliseconds > 0 {
		return LockedError(time.Duration(lockMilliseconds) * time.Millisecond)
	}
	return nil
}

func (guard *Guard) Reset(ctx context.Context, emailAddress string) {
	if err := guard.client.Del(ctx, guard.failureKey(emailAddress)).Err(); err != nil {
		guard.logUnavailable(ctx, err)
	}
}

func (guard *Guard) failureKey(emailAddress string) string {
	return guard.keyPrefix + "failures:" + hashEmail(emailAddress)
}

func (guard *Guard) lockKey(emailAddress string) string {
	return guard.keyPrefix + "lock:" + hashEmail(emailAddress)
}

func (guard *Guard) logUnavailable(ctx context.Context, err error) {
	if errors.Is(err, context.Canceled) {
		return
	}
	guard.logger.ErrorContext(ctx, "login lockout unavailable, allowing request", slog.String("error", err.Error()))
}

func hashEmail(emailAddress string) string {
	emailHash := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(emailAddress))))
	return hex.EncodeToString(emailHash[:])
}
