package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func newRequestIDTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.GET("/echo", func(context *gin.Context) {
		context.String(http.StatusOK, RequestIDFrom(context))
	})
	return router
}

func TestRequestIDGeneratesIDWhenHeaderIsMissing(t *testing.T) {
	recorder := httptest.NewRecorder()
	newRequestIDTestRouter().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/echo", nil))

	generatedID := recorder.Header().Get(RequestIDHeader)
	if _, err := uuid.Parse(generatedID); err != nil {
		t.Fatalf("expected a UUID request id, got %q", generatedID)
	}
	if recorder.Body.String() != generatedID {
		t.Fatalf("context request id %q does not match header %q", recorder.Body.String(), generatedID)
	}
}

func TestRequestIDKeepsValidIncomingHeader(t *testing.T) {
	incomingID := "client-trace-12345"
	request := httptest.NewRequest(http.MethodGet, "/echo", nil)
	request.Header.Set(RequestIDHeader, incomingID)
	recorder := httptest.NewRecorder()

	newRequestIDTestRouter().ServeHTTP(recorder, request)

	if recorder.Header().Get(RequestIDHeader) != incomingID {
		t.Fatalf("got %q, want %q", recorder.Header().Get(RequestIDHeader), incomingID)
	}
}

func TestRequestIDReplacesUnsafeIncomingHeader(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/echo", nil)
	request.Header.Set(RequestIDHeader, "bad id with spaces\nand newline")
	recorder := httptest.NewRecorder()

	newRequestIDTestRouter().ServeHTTP(recorder, request)

	if _, err := uuid.Parse(recorder.Header().Get(RequestIDHeader)); err != nil {
		t.Fatalf("expected unsafe header to be replaced with a UUID, got %q", recorder.Header().Get(RequestIDHeader))
	}
}
