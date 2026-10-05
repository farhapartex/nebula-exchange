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

	"github.com/farhapartex/nebula-exchange/backend/internal/health"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/logger"
)

type panickingRoutes struct{}

func (panickingRoutes) RegisterRoutes(router gin.IRouter) {
	router.GET("/explode", func(*gin.Context) { panic("boom") })
}

func newTestRouter(t *testing.T, logOutput *bytes.Buffer, registrars ...RouteRegistrar) *gin.Engine {
	t.Helper()
	testLogger := logger.NewWithWriter(logOutput, slog.LevelInfo, true)
	healthHandler := health.NewHandler(health.NewService(testLogger))
	router, err := NewRouter(RouterOptions{Logger: testLogger}, append([]RouteRegistrar{healthHandler}, registrars...)...)
	if err != nil {
		t.Fatalf("build router: %v", err)
	}
	return router
}

func decodeErrorCode(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()
	var responseBody struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	return responseBody.Error.Code
}

func TestHealthEndpointReturnsDataEnvelopeAndLogsRequestID(t *testing.T) {
	var logOutput bytes.Buffer
	router := newTestRouter(t, &logOutput)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))

	var responseBody struct {
		Data health.Report `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if recorder.Code != http.StatusOK || responseBody.Data.Status != health.StatusOK {
		t.Fatalf("got %d %+v, want 200 ok", recorder.Code, responseBody.Data)
	}
	requestID := recorder.Header().Get("X-Request-ID")
	if requestID == "" || !strings.Contains(logOutput.String(), requestID) {
		t.Fatalf("expected a log line with request id %q, got %s", requestID, logOutput.String())
	}
}

func TestPanicReturnsErrorEnvelope(t *testing.T) {
	router := newTestRouter(t, &bytes.Buffer{}, panickingRoutes{})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/explode", nil))

	if recorder.Code != http.StatusInternalServerError || decodeErrorCode(t, recorder) != "INTERNAL_ERROR" {
		t.Fatalf("got %d %s, want 500 INTERNAL_ERROR", recorder.Code, recorder.Body.String())
	}
}

func TestUnknownRoutesAndMethodsReturnErrorEnvelope(t *testing.T) {
	router := newTestRouter(t, &bytes.Buffer{})
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
			testRequest := httptest.NewRequest(expectation.method, expectation.path, nil)
			testRequest.Header.Set("X-Nebula-Client", "web")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, testRequest)
			if recorder.Code != expectation.expectedStatus || decodeErrorCode(t, recorder) != expectation.expectedCode {
				t.Fatalf("got %d %s, want %d %s", recorder.Code, recorder.Body.String(), expectation.expectedStatus, expectation.expectedCode)
			}
		})
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
	server := New(freeAddress, newTestRouter(t, &bytes.Buffer{}), testLogger)
	shutdownSignal, triggerShutdown := context.WithCancel(context.Background())
	runResult := make(chan error, 1)
	go func() { runResult <- server.Run(shutdownSignal, 2*time.Second) }()

	waitForServer(t, "http://"+freeAddress+"/api/v1/health")
	triggerShutdown()

	select {
	case err := <-runResult:
		if err != nil {
			t.Fatalf("expected a clean shutdown, got %v", err)
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
