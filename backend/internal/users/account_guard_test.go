package users_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/users"
)

type guardedRoutes struct {
	guard *users.AccountGuard
}

func (routes guardedRoutes) RegisterRoutes(router gin.IRouter) {
	respondOK := func(context *gin.Context) { context.Status(http.StatusNoContent) }
	router.POST("/actions", routes.guard.RequireStatus(users.StatusActive), respondOK)
	router.GET("/views", routes.guard.RequireStatus(users.StatusActive, users.StatusFrozen, users.StatusPendingPayment), respondOK)
	router.GET("/admin", routes.guard.RequireAdmin(), respondOK)
}

type guardTestHarness struct {
	pool         *pgxpool.Pool
	router       http.Handler
	accessTokens *accesstoken.Manager
}

func newGuardTestHarness(t *testing.T) *guardTestHarness {
	t.Helper()
	pool := databasetest.NewPool(t)
	accessTokens, _ := accesstoken.NewManager("guard-test-secret-with-at-least-32-chars", accesstoken.DefaultLifetime, time.Now)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	router := httpserver.NewRouter(
		httpserver.RouterOptions{Logger: testLogger, IdentifyUser: authentication.IdentifyUser(accessTokens)},
		guardedRoutes{guard: users.NewAccountGuard(pool, users.NewRepository())},
	)
	return &guardTestHarness{pool: pool, router: router, accessTokens: accessTokens}
}

func (harness *guardTestHarness) userWith(t *testing.T, status users.Status, isAdmin bool) string {
	t.Helper()
	createdUser, err := users.NewRepository().Create(context.Background(), harness.pool, users.NewUser{
		Email:           uuid.NewString() + "@nebula.test",
		Username:        "pilot_" + uuid.NewString()[:8],
		PasswordHash:    "unused",
		TermsAcceptedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	_, err = harness.pool.Exec(context.Background(),
		"UPDATE users SET is_active = true, activated_at = now(), status = $2, is_admin = $3 WHERE id = $1",
		createdUser.ID, string(status), isAdmin)
	if err != nil {
		t.Fatalf("set user state: %v", err)
	}
	issuedToken, _ := harness.accessTokens.Issue(createdUser.ID)
	return issuedToken.Value
}

func (harness *guardTestHarness) call(method, path, accessToken string) (int, string) {
	request := httptest.NewRequest(method, "/api/v1"+path, nil)
	request.Header.Set("X-Nebula-Client", "web")
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)
	var errorBody struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(recorder.Body.Bytes(), &errorBody)
	return recorder.Code, errorBody.Error.Code
}

func TestRequireStatusAllowsOnlyTheListedStatuses(t *testing.T) {
	harness := newGuardTestHarness(t)
	activeToken := harness.userWith(t, users.StatusActive, false)
	pendingToken := harness.userWith(t, users.StatusPendingPayment, false)
	frozenToken := harness.userWith(t, users.StatusFrozen, false)

	expectations := []struct {
		name           string
		method, path   string
		accessToken    string
		expectedStatus int
		expectedCode   string
	}{
		{"active may act", http.MethodPost, "/actions", activeToken, http.StatusNoContent, ""},
		{"pending may not act", http.MethodPost, "/actions", pendingToken, http.StatusForbidden, "ACCOUNT_NOT_ACTIVE"},
		{"frozen may not act", http.MethodPost, "/actions", frozenToken, http.StatusForbidden, "ACCOUNT_NOT_ACTIVE"},
		{"frozen may view", http.MethodGet, "/views", frozenToken, http.StatusNoContent, ""},
		{"anonymous is rejected", http.MethodGet, "/views", "", http.StatusUnauthorized, "UNAUTHORIZED"},
	}
	for _, expectation := range expectations {
		statusCode, errorCode := harness.call(expectation.method, expectation.path, expectation.accessToken)
		if statusCode != expectation.expectedStatus || errorCode != expectation.expectedCode {
			t.Fatalf("%s: got %d %q, want %d %q", expectation.name, statusCode, errorCode, expectation.expectedStatus, expectation.expectedCode)
		}
	}
}

func TestRequireAdminRejectsRegularPlayers(t *testing.T) {
	harness := newGuardTestHarness(t)

	if statusCode, _ := harness.call(http.MethodGet, "/admin", harness.userWith(t, users.StatusActive, false)); statusCode != http.StatusForbidden {
		t.Fatalf("regular player: got %d, want 403", statusCode)
	}
	if statusCode, _ := harness.call(http.MethodGet, "/admin", harness.userWith(t, users.StatusActive, true)); statusCode != http.StatusNoContent {
		t.Fatalf("admin: got %d, want 204", statusCode)
	}
}
