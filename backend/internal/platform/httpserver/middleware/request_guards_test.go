package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newGuardedRouter(isProduction bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(
		SecurityHeaders(isProduction),
		CrossOrigin([]string{"http://localhost:3000"}),
		RequireClientIdentification("/webhooks/"),
	)
	respondOK := func(context *gin.Context) { context.Status(http.StatusOK) }
	router.GET("/orders", respondOK)
	router.POST("/orders", respondOK)
	router.POST("/webhooks/stripe", respondOK)
	return router
}

func serve(router *gin.Engine, request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestClientIdentificationIsRequiredForStateChangingRequests(t *testing.T) {
	router := newGuardedRouter(false)

	missingHeaderResponse := serve(router, httptest.NewRequest(http.MethodPost, "/orders", nil))

	identifiedRequest := httptest.NewRequest(http.MethodPost, "/orders", nil)
	identifiedRequest.Header.Set(ClientIdentificationHeader, WebClientIdentifier)
	identifiedResponse := serve(router, identifiedRequest)

	wrongValueRequest := httptest.NewRequest(http.MethodPost, "/orders", nil)
	wrongValueRequest.Header.Set(ClientIdentificationHeader, "curl")
	wrongValueResponse := serve(router, wrongValueRequest)

	if missingHeaderResponse.Code != http.StatusForbidden || wrongValueResponse.Code != http.StatusForbidden {
		t.Fatalf("got %d and %d, want 403 without a valid header", missingHeaderResponse.Code, wrongValueResponse.Code)
	}
	if identifiedResponse.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 with the header", identifiedResponse.Code)
	}
}

func TestClientIdentificationSkipsSafeMethodsAndWebhooks(t *testing.T) {
	router := newGuardedRouter(false)

	readResponse := serve(router, httptest.NewRequest(http.MethodGet, "/orders", nil))
	webhookResponse := serve(router, httptest.NewRequest(http.MethodPost, "/webhooks/stripe", nil))

	if readResponse.Code != http.StatusOK || webhookResponse.Code != http.StatusOK {
		t.Fatalf("got %d and %d, want both 200", readResponse.Code, webhookResponse.Code)
	}
}

func TestCrossOriginAllowsOnlyTheFrontendOrigin(t *testing.T) {
	router := newGuardedRouter(false)

	allowedPreflight := httptest.NewRequest(http.MethodOptions, "/orders", nil)
	allowedPreflight.Header.Set("Origin", "http://localhost:3000")
	allowedPreflight.Header.Set("Access-Control-Request-Method", http.MethodPost)
	allowedPreflight.Header.Set("Access-Control-Request-Headers", "content-type,idempotency-key,x-nebula-client")
	allowedResponse := serve(router, allowedPreflight)

	foreignRequest := httptest.NewRequest(http.MethodGet, "/orders", nil)
	foreignRequest.Header.Set("Origin", "https://evil.example")
	foreignResponse := serve(router, foreignRequest)

	if allowedResponse.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("expected frontend origin to be allowed, headers %v", allowedResponse.Header())
	}
	if allowedResponse.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatal("expected credentials to be allowed for the refresh cookie")
	}
	if foreignResponse.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("foreign origin must not be allowed")
	}
}

func TestSecurityHeadersAreSet(t *testing.T) {
	developmentResponse := serve(newGuardedRouter(false), httptest.NewRequest(http.MethodGet, "/orders", nil))
	productionResponse := serve(newGuardedRouter(true), httptest.NewRequest(http.MethodGet, "/orders", nil))

	if developmentResponse.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("expected nosniff header")
	}
	if developmentResponse.Header().Get("Strict-Transport-Security") != "" {
		t.Fatal("HSTS must not be sent outside production")
	}
	if productionResponse.Header().Get("Strict-Transport-Security") == "" {
		t.Fatal("expected HSTS in production")
	}
}
