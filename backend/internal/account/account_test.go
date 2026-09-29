package account_test

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

	"nebula-exchange/backend/internal/account"
	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/auth/passwordhash"
	"nebula-exchange/backend/internal/auth/session"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/users"
)

const currentPassword = "Mining4Crystal!Moon"

var fastHashParameters = passwordhash.Parameters{MemoryInKibibytes: 1024, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}

type accountHarness struct {
	pool          *pgxpool.Pool
	router        http.Handler
	accessTokens  *accesstoken.Manager
	refreshTokens *session.RefreshTokens
	repository    *users.Repository
}

func allowEverything(context *gin.Context) { context.Next() }

func newAccountHarness(t *testing.T) *accountHarness {
	t.Helper()
	pool := databasetest.NewPool(t)
	accessTokens, _ := accesstoken.NewManager("account-test-secret-with-at-least-32-chars", accesstoken.DefaultLifetime, time.Now)
	repository := users.NewRepository()
	activeSessions := session.NewSessions(pool, time.Now)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	accountService := account.NewService(account.Dependencies{
		Pool:           pool,
		Users:          repository,
		PasswordHasher: passwordhash.NewHasher(passwordhash.HasherOptions{Parameters: fastHashParameters}),
		Sessions:       activeSessions,
		Now:            time.Now,
	})
	router := httpserver.NewRouter(
		httpserver.RouterOptions{Logger: testLogger, IdentifyUser: authentication.IdentifyUser(accessTokens)},
		account.NewHandler(accountService, allowEverything),
		session.NewSessionsHandler(activeSessions, session.CookieSettings{IsSecure: true}),
	)
	return &accountHarness{
		pool:          pool,
		router:        router,
		accessTokens:  accessTokens,
		refreshTokens: session.NewRefreshTokens(session.DefaultRefreshTokenLifetime, time.Now),
		repository:    repository,
	}
}

type signedInDevice struct {
	accessToken  string
	refreshToken session.IssuedRefreshToken
}

func (harness *accountHarness) createUserWithDevices(t *testing.T, username string, deviceCount int) (uuid.UUID, []signedInDevice) {
	t.Helper()
	passwordHash, _ := passwordhash.Hash(currentPassword, fastHashParameters)
	createdUser, err := harness.repository.Create(context.Background(), harness.pool, users.NewUser{
		Email:           username + "@nebula.test",
		Username:        username,
		PasswordHash:    passwordHash,
		TermsAcceptedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	devices := make([]signedInDevice, 0, deviceCount)
	for deviceIndex := 0; deviceIndex < deviceCount; deviceIndex++ {
		refreshToken, err := harness.refreshTokens.Issue(context.Background(), harness.pool, createdUser.ID, session.ClientMetadata{
			UserAgent: "Browser " + string(rune('A'+deviceIndex)),
			IPAddress: "203.0.113.5",
		})
		if err != nil {
			t.Fatalf("issue refresh token: %v", err)
		}
		accessToken, _ := harness.accessTokens.Issue(createdUser.ID)
		devices = append(devices, signedInDevice{accessToken: accessToken.Value, refreshToken: refreshToken})
		time.Sleep(5 * time.Millisecond)
	}
	return createdUser.ID, devices
}

func (harness *accountHarness) send(t *testing.T, method, path string, device signedInDevice, requestBody any) *httptest.ResponseRecorder {
	t.Helper()
	var encodedBody []byte
	if requestBody != nil {
		encodedBody, _ = json.Marshal(requestBody)
	}
	request := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(encodedBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Nebula-Client", "web")
	request.Header.Set("Authorization", "Bearer "+device.accessToken)
	request.AddCookie(&http.Cookie{Name: session.RefreshCookieName, Value: device.refreshToken.Plaintext})
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)
	return recorder
}

func fieldError(t *testing.T, recorder *httptest.ResponseRecorder, fieldName string) string {
	t.Helper()
	var errorBody struct {
		Error struct {
			Details map[string]string `json:"details"`
		} `json:"error"`
	}
	_ = json.Unmarshal(recorder.Body.Bytes(), &errorBody)
	return errorBody.Error.Details[fieldName]
}

func (harness *accountHarness) isRevoked(t *testing.T, refreshToken session.IssuedRefreshToken) bool {
	t.Helper()
	var revokedAt *time.Time
	_ = harness.pool.QueryRow(context.Background(), "SELECT revoked_at FROM refresh_tokens WHERE id = $1", refreshToken.ID).Scan(&revokedAt)
	return revokedAt != nil
}

func TestUsernameCanBeChangedWhenValidAndFree(t *testing.T) {
	harness := newAccountHarness(t)
	_, devices := harness.createUserWithDevices(t, "pilot_one", 1)
	harness.createUserWithDevices(t, "pilot_two", 1)

	if recorder := harness.send(t, http.MethodPatch, "/me", devices[0], map[string]string{"username": "Star_Runner"}); recorder.Code != http.StatusOK || !bytes.Contains(recorder.Body.Bytes(), []byte("Star_Runner")) {
		t.Fatalf("rename: got %d %s", recorder.Code, recorder.Body.String())
	}
	if recorder := harness.send(t, http.MethodPatch, "/me", devices[0], map[string]string{"username": "PILOT_TWO"}); fieldError(t, recorder, "username") != "is already taken" {
		t.Fatalf("taken name: got %d %s", recorder.Code, recorder.Body.String())
	}
	if recorder := harness.send(t, http.MethodPatch, "/me", devices[0], map[string]string{"username": "no spaces"}); recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid name: got %d", recorder.Code)
	}
}

func TestPasswordChangeKeepsThisSessionAndSignsOutOthers(t *testing.T) {
	harness := newAccountHarness(t)
	userID, devices := harness.createUserWithDevices(t, "pilot_one", 3)

	wrongRecorder := harness.send(t, http.MethodPost, "/auth/password-changes", devices[0], map[string]string{"current_password": "not-my-password", "new_password": "Brand4New!Galaxy"})
	if fieldError(t, wrongRecorder, "current_password") != "is incorrect" {
		t.Fatalf("wrong current password: got %d %s", wrongRecorder.Code, wrongRecorder.Body.String())
	}

	recorder := harness.send(t, http.MethodPost, "/auth/password-changes", devices[0], map[string]string{"current_password": currentPassword, "new_password": "Brand4New!Galaxy"})
	if recorder.Code != http.StatusOK || !bytes.Contains(recorder.Body.Bytes(), []byte(`"signed_out_other_sessions":2`)) {
		t.Fatalf("change: got %d %s", recorder.Code, recorder.Body.String())
	}
	if harness.isRevoked(t, devices[0].refreshToken) {
		t.Fatal("the session that changed the password must stay signed in")
	}
	if !harness.isRevoked(t, devices[1].refreshToken) || !harness.isRevoked(t, devices[2].refreshToken) {
		t.Fatal("other sessions must be signed out")
	}
	storedHash, _ := harness.repository.FindPasswordHash(context.Background(), harness.pool, userID)
	if isMatch, _ := passwordhash.Verify("Brand4New!Galaxy", storedHash); !isMatch {
		t.Fatal("the new password must be stored")
	}
}

func TestSessionsAreListedWithPaginationAndCanBeRevoked(t *testing.T) {
	harness := newAccountHarness(t)
	_, devices := harness.createUserWithDevices(t, "pilot_one", 2)

	var firstPage struct {
		Data []struct {
			ID        string `json:"id"`
			UserAgent string `json:"user_agent"`
			IsCurrent bool   `json:"is_current"`
		} `json:"data"`
		Pagination struct {
			NextCursor *string `json:"next_cursor"`
		} `json:"pagination"`
	}
	firstPageRecorder := harness.send(t, http.MethodGet, "/auth/sessions?limit=1", devices[0], nil)
	_ = json.Unmarshal(firstPageRecorder.Body.Bytes(), &firstPage)
	if firstPageRecorder.Code != http.StatusOK || len(firstPage.Data) != 1 || firstPage.Pagination.NextCursor == nil {
		t.Fatalf("first page: got %d %s", firstPageRecorder.Code, firstPageRecorder.Body.String())
	}
	if firstPage.Data[0].UserAgent != "Browser B" || firstPage.Data[0].IsCurrent {
		t.Fatalf("newest session should be listed first and not be the current one, got %+v", firstPage.Data[0])
	}

	secondPageRecorder := harness.send(t, http.MethodGet, "/auth/sessions?limit=1&cursor="+*firstPage.Pagination.NextCursor, devices[0], nil)
	if !bytes.Contains(secondPageRecorder.Body.Bytes(), []byte(`"is_current":true`)) {
		t.Fatalf("second page should contain the current session, got %s", secondPageRecorder.Body.String())
	}

	otherRevokeRecorder := harness.send(t, http.MethodDelete, "/auth/sessions/"+firstPage.Data[0].ID, devices[0], nil)
	if otherRevokeRecorder.Code != http.StatusOK || !harness.isRevoked(t, devices[1].refreshToken) {
		t.Fatalf("revoke other: got %d %s", otherRevokeRecorder.Code, otherRevokeRecorder.Body.String())
	}
	currentRevokeRecorder := harness.send(t, http.MethodDelete, "/auth/sessions/"+devices[0].refreshToken.ID.String(), devices[0], nil)
	if !bytes.Contains(currentRevokeRecorder.Body.Bytes(), []byte(`"was_current_session":true`)) {
		t.Fatalf("revoke current: got %s", currentRevokeRecorder.Body.String())
	}
	if missingRecorder := harness.send(t, http.MethodDelete, "/auth/sessions/"+uuid.NewString(), devices[0], nil); missingRecorder.Code != http.StatusNotFound {
		t.Fatalf("unknown session: got %d", missingRecorder.Code)
	}
}
