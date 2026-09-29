package ratelimit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/redis/go-redis/v9"
)

type Policy struct {
	Name   string
	Limit  int
	Window time.Duration
}

type Decision struct {
	IsAllowed  bool
	Remaining  int
	RetryAfter time.Duration
}

var slidingWindowScript = redis.NewScript(`
local windowStart = tonumber(ARGV[1]) - tonumber(ARGV[2])
redis.call("ZREMRANGEBYSCORE", KEYS[1], "-inf", windowStart)
local requestCount = redis.call("ZCARD", KEYS[1])
if requestCount < tonumber(ARGV[3]) then
  redis.call("ZADD", KEYS[1], ARGV[1], ARGV[4])
  redis.call("PEXPIRE", KEYS[1], ARGV[2])
  return {1, tonumber(ARGV[3]) - requestCount - 1, 0}
end
local oldestEntry = redis.call("ZRANGE", KEYS[1], 0, 0, "WITHSCORES")
return {0, 0, tonumber(oldestEntry[2]) + tonumber(ARGV[2]) - tonumber(ARGV[1])}
`)

type SlidingWindowLimiter struct {
	client    redis.Scripter
	keyPrefix string
	now       func() time.Time
}

func NewSlidingWindowLimiter(client redis.Scripter, keyPrefix string, now func() time.Time) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{client: client, keyPrefix: keyPrefix, now: now}
}

func (limiter *SlidingWindowLimiter) Allow(ctx context.Context, policy Policy, subject string) (Decision, error) {
	nowInMilliseconds := limiter.now().UnixMilli()
	requestMember, err := uniqueRequestMember(nowInMilliseconds)
	if err != nil {
		return Decision{}, err
	}

	scriptResult, err := slidingWindowScript.Run(ctx, limiter.client,
		[]string{limiter.keyPrefix + policy.Name + ":" + subject},
		nowInMilliseconds, policy.Window.Milliseconds(), policy.Limit, requestMember,
	).Int64Slice()
	if err != nil {
		return Decision{}, err
	}

	return Decision{
		IsAllowed:  scriptResult[0] == 1,
		Remaining:  int(scriptResult[1]),
		RetryAfter: time.Duration(scriptResult[2]) * time.Millisecond,
	}, nil
}

func uniqueRequestMember(nowInMilliseconds int64) (string, error) {
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return time.UnixMilli(nowInMilliseconds).Format("150405.000") + ":" + hex.EncodeToString(randomBytes), nil
}
