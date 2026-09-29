package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/health"
	"nebula-exchange/backend/internal/platform/logger"
)

type panickingRoutes struct{}

func (panickingRoutes) RegisterRoutes(router gin.IRouter) {
	router.GET("/explode", func(*gin.Context) { panic("boom") })
}

func TestHealthEndpointReturnsDataEnvelopeAndLogsRequestID(t *testing.T) {
	var logOutput bytes.Buffer
	testLogger := logger.NewWithWriter(&logOutput, slog.LevelInfo, true)
	router := NewRouter(RouterOptions{Logger: testLogger}, health.NewHandler())

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", recorder.Code)
	}

	var responseBody struct {
		Data health.Status `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if responseBody.Data.Status != "ok" {
		t.Fatalf("got status %q, want ok", responseBody.Data.Status)
	}

	requestID := recorder.Header().Get("X-Request-ID")
	if requestID == "" || !strings.Contains(logOutput.String(), requestID) {
		t.Fatalf("expected log line with request id %q, got %s", requestID, logOutput.String())
	}
}

func TestPanicReturnsErrorEnvelope(t *testing.T) {
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelInfo, true)
	router := NewRouter(RouterOptions{Logger: testLogger}, panickingRoutes{})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/explode", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("got status %d, want 500", recorder.Code)
	}
	var responseBody struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if responseBody.Error.Code != "INTERNAL_ERROR" {
		t.Fatalf("got error code %q, want INTERNAL_ERROR", responseBody.Error.Code)
	}
}

func TestRunStopsGracefullyWhenShutdownIsSignalled(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	freeAddress := listener.Addr().String()
	listener.Close()

	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelInfo, true)
	server := New(freeAddress, NewRouter(RouterOptions{Logger: testLogger}, health.NewHandler()), testLogger)

	shutdownSignal, triggerShutdown := context.WithCancel(context.Background())
	runResult := make(chan error, 1)
	go func() { runResult <- server.Run(shutdownSignal, 2*time.Second) }()

	waitForServer(t, "http://"+freeAddress+"/api/v1/health")
	triggerShutdown()

	select {
	case err := <-runResult:
		if err != nil {
			t.Fatalf("expected clean shutdown, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server did not stop within the shutdown timeout")
	}
}

func waitForServer(t *testing.T, healthURL string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		healthResponse, err := http.Get(healthURL)
		if err == nil {
			healthResponse.Body.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("server at %s did not become ready", healthURL)
}

func TestUnknownRoutesAndMethodsReturnErrorEnvelope(t *testing.T) {
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelInfo, true)
	router := NewRouter(RouterOptions{Logger: testLogger}, health.NewHandler())

	expectations := map[string]struct {
		method         string
		path           string
		expectedStatus int
		expectedCode   string
	}{
		"unknown route":  {method: http.MethodGet, path: "/api/v1/missing", expectedStatus: http.StatusNotFound, expectedCode: "NOT_FOUND"},
		"unknown method": {method: http.MethodPost, path: "/api/v1/health", expectedStatus: http.StatusMethodNotAllowed, expectedCode: "METHOD_NOT_ALLOWED"},
	}

	for caseName, expectation := range expectations {
		t.Run(caseName, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			testRequest := httptest.NewRequest(expectation.method, expectation.path, nil)
			testRequest.Header.Set("X-Nebula-Client", "web")
			router.ServeHTTP(recorder, testRequest)

			var responseBody struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &responseBody); err != nil {
				t.Fatalf("response is not valid JSON: %v", err)
			}
			if recorder.Code != expectation.expectedStatus || responseBody.Error.Code != expectation.expectedCode {
				t.Fatalf("got %d %s, want %d %s", recorder.Code, responseBody.Error.Code, expectation.expectedStatus, expectation.expectedCode)
			}
		})
	}
}
