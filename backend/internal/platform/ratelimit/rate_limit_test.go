package ratelimit_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/platform/ratelimit"
	"nebula-exchange/backend/internal/platform/redisclient/redistest"
)

type movableClock struct {
	currentTime atomic.Pointer[time.Time]
}

func (clock *movableClock) now() time.Time { return *clock.currentTime.Load() }
func (clock *movableClock) advance(duration time.Duration) {
	advancedTime := clock.now().Add(duration)
	clock.currentTime.Store(&advancedTime)
}

func newLimiter(t *testing.T) (*ratelimit.SlidingWindowLimiter, *movableClock) {
	t.Helper()
	clock := &movableClock{}
	startTime := time.Now()
	clock.currentTime.Store(&startTime)
	keyPrefix := "test:ratelimit:" + uuid.NewString() + ":"
	return ratelimit.NewSlidingWindowLimiter(redistest.NewClient(t), keyPrefix, clock.now), clock
}

var testPolicy = ratelimit.Policy{Name: "signup", Limit: 3, Window: time.Minute}

func TestLimiterAllowsUpToTheLimitThenBlocks(t *testing.T) {
	limiter, _ := newLimiter(t)

	for requestNumber := 1; requestNumber <= 3; requestNumber++ {
		decision, err := limiter.Allow(context.Background(), testPolicy, "203.0.113.7")
		if err != nil || !decision.IsAllowed || decision.Remaining != 3-requestNumber {
			t.Fatalf("request %d: got %+v, %v", requestNumber, decision, err)
		}
	}

	blockedDecision, err := limiter.Allow(context.Background(), testPolicy, "203.0.113.7")
	if err != nil || blockedDecision.IsAllowed {
		t.Fatalf("fourth request should be blocked, got %+v, %v", blockedDecision, err)
	}
	if blockedDecision.RetryAfter <= 0 || blockedDecision.RetryAfter > time.Minute {
		t.Fatalf("got retry after %v", blockedDecision.RetryAfter)
	}
}

func TestLimiterSlidesTheWindowAndKeepsSubjectsSeparate(t *testing.T) {
	limiter, clock := newLimiter(t)

	for requestNumber := 0; requestNumber < 3; requestNumber++ {
		limiter.Allow(context.Background(), testPolicy, "203.0.113.7")
		clock.advance(10 * time.Second)
	}
	if decision, _ := limiter.Allow(context.Background(), testPolicy, "203.0.113.7"); decision.IsAllowed {
		t.Fatal("the window is still full")
	}
	if decision, _ := limiter.Allow(context.Background(), testPolicy, "198.51.100.4"); !decision.IsAllowed {
		t.Fatal("another IP must have its own window")
	}
	loginPolicy := ratelimit.Policy{Name: "login", Limit: 3, Window: time.Minute}
	if decision, _ := limiter.Allow(context.Background(), loginPolicy, "203.0.113.7"); !decision.IsAllowed {
		t.Fatal("another policy must have its own window")
	}

	clock.advance(31 * time.Second)
	if decision, _ := limiter.Allow(context.Background(), testPolicy, "203.0.113.7"); !decision.IsAllowed {
		t.Fatal("the oldest request left the window, so one more must be allowed")
	}
}

func TestMiddlewareRespondsWithRateLimitedAndRetryAfter(t *testing.T) {
	limiter, _ := newLimiter(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	router.POST("/auth/signup", ratelimit.NewMiddlewareFactory(limiter, testLogger).PerClientIP(testPolicy), func(context *gin.Context) {
		context.Status(http.StatusCreated)
	})

	var lastRecorder *httptest.ResponseRecorder
	for requestNumber := 0; requestNumber < 4; requestNumber++ {
		lastRecorder = httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/auth/signup", nil)
		request.RemoteAddr = "203.0.113.7:51234"
		router.ServeHTTP(lastRecorder, request)
	}

	if lastRecorder.Code != http.StatusTooManyRequests || lastRecorder.Header().Get("Retry-After") == "" {
		t.Fatalf("got %d with headers %v", lastRecorder.Code, lastRecorder.Header())
	}
	var errorBody struct {
		Error struct {
			Code    string         `json:"code"`
			Details map[string]int `json:"details"`
		} `json:"error"`
	}
	_ = json.Unmarshal(lastRecorder.Body.Bytes(), &errorBody)
	if errorBody.Error.Code != "RATE_LIMITED" || errorBody.Error.Details["retry_after_seconds"] < 1 {
		t.Fatalf("unexpected body %s", lastRecorder.Body.String())
	}
}
