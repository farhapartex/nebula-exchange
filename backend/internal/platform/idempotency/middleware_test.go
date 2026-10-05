package idempotency_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database/databasetest"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/middleware"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/idempotency"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/logger"
)

type idempotencyTestHarness struct {
	router            *gin.Engine
	store             *idempotency.GormStore
	handlerCallCount  atomic.Int32
	nextStatusCode    atomic.Int32
	shouldPanicOnCall atomic.Bool
}

func newIdempotencyTestHarness(t *testing.T, options idempotency.Options) *idempotencyTestHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)

	harness := &idempotencyTestHarness{store: idempotency.NewGormStore(databasetest.Open(t))}
	harness.nextStatusCode.Store(http.StatusCreated)

	router := gin.New()
	router.Use(middleware.PanicRecovery(testLogger), idempotency.Middleware(harness.store, testLogger, options))
	handleOrder := func(context *gin.Context) {
		callNumber := harness.handlerCallCount.Add(1)
		if harness.shouldPanicOnCall.Load() {
			panic("handler exploded")
		}
		context.JSON(int(harness.nextStatusCode.Load()), gin.H{"data": gin.H{"call_number": callNumber}})
	}
	router.POST("/orders", handleOrder)
	router.GET("/orders", handleOrder)
	harness.router = router
	return harness
}

func (harness *idempotencyTestHarness) send(method, key, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, "/orders", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if key != "" {
		request.Header.Set(idempotency.KeyHeader, key)
	}
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)
	return recorder
}

func requireStatus(t *testing.T, recorder *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	if recorder.Code != expectedStatus {
		t.Fatalf("got status %d, want %d, body %s", recorder.Code, expectedStatus, recorder.Body.String())
	}
}

const orderBody = `{"tool":"iron_pipe","price":"25000000"}`

func TestRepeatedKeyReplaysSuccessfulResponse(t *testing.T) {
	harness := newIdempotencyTestHarness(t, idempotency.Options{})

	firstResponse := harness.send(http.MethodPost, "order-key-0001", orderBody)
	secondResponse := harness.send(http.MethodPost, "order-key-0001", orderBody)

	requireStatus(t, firstResponse, http.StatusCreated)
	requireStatus(t, secondResponse, http.StatusCreated)
	if secondResponse.Body.String() != firstResponse.Body.String() {
		t.Fatalf("replayed body %s differs from original %s", secondResponse.Body.String(), firstResponse.Body.String())
	}
	if secondResponse.Header().Get(idempotency.ReplayedHeader) != "true" {
		t.Fatal("expected replayed response to be marked")
	}
	if harness.handlerCallCount.Load() != 1 {
		t.Fatalf("handler ran %d times, want 1", harness.handlerCallCount.Load())
	}
}

func TestRepeatedKeyReplaysClientErrors(t *testing.T) {
	harness := newIdempotencyTestHarness(t, idempotency.Options{})
	harness.nextStatusCode.Store(http.StatusUnprocessableEntity)

	harness.send(http.MethodPost, "order-key-0002", orderBody)
	harness.nextStatusCode.Store(http.StatusCreated)
	replayedResponse := harness.send(http.MethodPost, "order-key-0002", orderBody)

	requireStatus(t, replayedResponse, http.StatusUnprocessableEntity)
	if harness.handlerCallCount.Load() != 1 {
		t.Fatalf("handler ran %d times, want 1", harness.handlerCallCount.Load())
	}
}

func TestServerErrorsAreNotStored(t *testing.T) {
	harness := newIdempotencyTestHarness(t, idempotency.Options{})
	harness.nextStatusCode.Store(http.StatusServiceUnavailable)

	harness.send(http.MethodPost, "order-key-0003", orderBody)
	harness.nextStatusCode.Store(http.StatusCreated)
	retriedResponse := harness.send(http.MethodPost, "order-key-0003", orderBody)

	requireStatus(t, retriedResponse, http.StatusCreated)
	if harness.handlerCallCount.Load() != 2 {
		t.Fatalf("handler ran %d times, want 2", harness.handlerCallCount.Load())
	}
}

func TestSameKeyWithDifferentBodyConflicts(t *testing.T) {
	harness := newIdempotencyTestHarness(t, idempotency.Options{})

	harness.send(http.MethodPost, "order-key-0004", orderBody)
	conflictingResponse := harness.send(http.MethodPost, "order-key-0004", `{"tool":"iron_pipe","price":"26000000"}`)

	requireStatus(t, conflictingResponse, http.StatusConflict)
	if !strings.Contains(conflictingResponse.Body.String(), "different request") {
		t.Fatalf("unexpected body %s", conflictingResponse.Body.String())
	}
}

func TestKeyInProgressConflicts(t *testing.T) {
	harness := newIdempotencyTestHarness(t, idempotency.Options{})
	firstResponse := harness.send(http.MethodPost, "order-key-0005", orderBody)
	requireStatus(t, firstResponse, http.StatusCreated)

	if _, err := harness.store.Begin(context.Background(), "anonymous", "order-key-0006", []byte("pending")); err != nil {
		t.Fatalf("seed in-progress key: %v", err)
	}
	inProgressResponse := harness.send(http.MethodPost, "order-key-0006", orderBody)

	requireStatus(t, inProgressResponse, http.StatusConflict)
}

func TestExpiredKeyRunsAgain(t *testing.T) {
	farFuture := time.Now().Add(48 * time.Hour)
	harness := newIdempotencyTestHarness(t, idempotency.Options{Now: func() time.Time { return farFuture }})

	harness.send(http.MethodPost, "order-key-0007", orderBody)
	secondResponse := harness.send(http.MethodPost, "order-key-0007", orderBody)

	requireStatus(t, secondResponse, http.StatusCreated)
	if harness.handlerCallCount.Load() != 2 {
		t.Fatalf("handler ran %d times, want 2 after expiry", harness.handlerCallCount.Load())
	}
}

func TestStaleInProgressKeyIsReclaimed(t *testing.T) {
	afterInProgressTimeout := time.Now().Add(5 * time.Minute)
	harness := newIdempotencyTestHarness(t, idempotency.Options{Now: func() time.Time { return afterInProgressTimeout }})
	if _, err := harness.store.Begin(context.Background(), "anonymous", "order-key-0008", []byte("stale")); err != nil {
		t.Fatalf("seed stale key: %v", err)
	}

	reclaimedResponse := harness.send(http.MethodPost, "order-key-0008", orderBody)

	requireStatus(t, reclaimedResponse, http.StatusCreated)
}

func TestPanicReleasesKey(t *testing.T) {
	harness := newIdempotencyTestHarness(t, idempotency.Options{})
	harness.shouldPanicOnCall.Store(true)

	panicResponse := harness.send(http.MethodPost, "order-key-0009", orderBody)
	harness.shouldPanicOnCall.Store(false)
	retriedResponse := harness.send(http.MethodPost, "order-key-0009", orderBody)

	requireStatus(t, panicResponse, http.StatusInternalServerError)
	requireStatus(t, retriedResponse, http.StatusCreated)
}

func TestRequestsWithoutKeyOrWithSafeMethodAreNotTracked(t *testing.T) {
	harness := newIdempotencyTestHarness(t, idempotency.Options{})

	harness.send(http.MethodPost, "", orderBody)
	harness.send(http.MethodPost, "", orderBody)
	harness.send(http.MethodGet, "order-key-0010", "")
	harness.send(http.MethodGet, "order-key-0010", "")

	if harness.handlerCallCount.Load() != 4 {
		t.Fatalf("handler ran %d times, want 4", harness.handlerCallCount.Load())
	}
}

func TestMalformedKeyIsRejected(t *testing.T) {
	harness := newIdempotencyTestHarness(t, idempotency.Options{})

	rejectedResponse := harness.send(http.MethodPost, "short", orderBody)

	requireStatus(t, rejectedResponse, http.StatusBadRequest)
	if harness.handlerCallCount.Load() != 0 {
		t.Fatal("handler should not run for a malformed key")
	}
}
