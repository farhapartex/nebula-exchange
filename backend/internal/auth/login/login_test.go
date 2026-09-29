package login_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/auth/login"
	"nebula-exchange/backend/internal/auth/passwordhash"
	"nebula-exchange/backend/internal/auth/session"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/users"
)

const testPassword = "Mining4Crystal!Moon"

var fastHashParameters = passwordhash.Parameters{MemoryInKibibytes: 1024, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}

type loginTestHarness struct {
	pool         *pgxpool.Pool
	router       http.Handler
	accessTokens *accesstoken.Manager
	repository   *users.Repository
}

func newLoginTestHarness(t *testing.T) *loginTestHarness {
	t.Helper()
	pool := databasetest.NewPool(t)
	accessTokens, err := accesstoken.NewManager("login-test-secret-with-at-least-32-chars", accesstoken.DefaultLifetime, time.Now)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	repository := users.NewRepository()
	loginService, err := login.NewService(login.Dependencies{
		Pool:                pool,
		Users:               repository,
		AccessTokens:        accessTokens,
		RefreshTokens:       session.NewRefreshTokens(session.DefaultRefreshTokenLifetime, time.Now),
		PasswordHasher:      passwordhash.NewHasher(passwordhash.HasherOptions{Parameters: fastHashParameters}),
		PasswordHashOptions: fastHashParameters,
		Now:                 time.Now,
	})
	if err != nil {
		t.Fatalf("create login service: %v", err)
	}

	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	router := httpserver.NewRouter(
		httpserver.RouterOptions{Logger: testLogger, IdentifyUser: authentication.IdentifyUser(accessTokens)},
		login.NewHandler(loginService, session.CookieSettings{IsSecure: true}, time.Now),
		users.NewMeHandler(pool, repository),
	)
	return &loginTestHarness{pool: pool, router: router, accessTokens: accessTokens, repository: repository}
}

func (harness *loginTestHarness) createUser(t *testing.T, email string, isActive bool, status users.Status) uuid.UUID {
	t.Helper()
	passwordHash, _ := passwordhash.Hash(testPassword, fastHashParameters)
	createdUser, err := harness.repository.Create(context.Background(), harness.pool, users.NewUser{
		Email:           email,
		Username:        "pilot_" + uuid.NewString()[:8],
		PasswordHash:    passwordHash,
		TermsAcceptedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if isActive {
		_, err = harness.pool.Exec(context.Background(),
			"UPDATE users SET is_active = true, activated_at = now(), status = $2 WHERE id = $1", createdUser.ID, string(status))
		if err != nil {
			t.Fatalf("activate user: %v", err)
		}
	}
	return createdUser.ID
}

type sessionEnvelope struct {
	Data struct {
		AccessToken string        `json:"access_token"`
		User        users.Profile `json:"user"`
	} `json:"data"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (harness *loginTestHarness) send(t *testing.T, method, path string, requestBody any, modify func(*http.Request)) (*httptest.ResponseRecorder, sessionEnvelope) {
	t.Helper()
	var bodyReader *bytes.Reader
	if requestBody != nil {
		encodedBody, _ := json.Marshal(requestBody)
		bodyReader = bytes.NewReader(encodedBody)
	} else {
		bodyReader = bytes.NewReader(nil)
	}
	request := httptest.NewRequest(method, path, bodyReader)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Nebula-Client", "web")
	if modify != nil {
		modify(request)
	}
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)

	var decodedBody sessionEnvelope
	_ = json.Unmarshal(recorder.Body.Bytes(), &decodedBody)
	return recorder, decodedBody
}

func (harness *loginTestHarness) logIn(t *testing.T, email, password string) (*httptest.ResponseRecorder, sessionEnvelope) {
	t.Helper()
	return harness.send(t, http.MethodPost, "/api/v1/auth/login", map[string]string{"email": email, "password": password}, nil)
}

func refreshCookieFrom(t *testing.T, recorder *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, responseCookie := range recorder.Result().Cookies() {
		if responseCookie.Name == session.RefreshCookieName {
			return responseCookie
		}
	}
	t.Fatal("response has no refresh cookie")
	return nil
}

func TestLoginIssuesAccessTokenAndRefreshCookie(t *testing.T) {
	harness := newLoginTestHarness(t)
	userID := harness.createUser(t, "pilot@nebula.test", true, users.StatusPendingPayment)

	recorder, loggedIn := harness.logIn(t, "  PILOT@nebula.test ", testPassword)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, error %+v", recorder.Code, loggedIn.Error)
	}
	verifiedUserID, err := harness.accessTokens.Verify(loggedIn.Data.AccessToken)
	if err != nil || verifiedUserID != userID {
		t.Fatalf("access token does not verify for the user: %v", err)
	}
	if loggedIn.Data.User.Email != "pilot@nebula.test" || loggedIn.Data.User.LastLoginAt == nil {
		t.Fatalf("unexpected user %+v", loggedIn.Data.User)
	}

	refreshCookie := refreshCookieFrom(t, recorder)
	if !refreshCookie.HttpOnly || !refreshCookie.Secure || refreshCookie.SameSite != http.SameSiteStrictMode || refreshCookie.Path != "/api/v1/auth" {
		t.Fatalf("refresh cookie is not locked down: %+v", refreshCookie)
	}

	var storedLastLogin *time.Time
	_ = harness.pool.QueryRow(context.Background(), "SELECT last_login_at FROM users WHERE id = $1", userID).Scan(&storedLastLogin)
	if storedLastLogin == nil {
		t.Fatal("last_login_at was not recorded")
	}
}

func TestLoginRejectsWrongCredentialsWithTheSameMessage(t *testing.T) {
	harness := newLoginTestHarness(t)
	harness.createUser(t, "pilot@nebula.test", true, users.StatusActive)

	wrongPasswordRecorder, wrongPassword := harness.logIn(t, "pilot@nebula.test", "not-the-password")
	unknownEmailRecorder, unknownEmail := harness.logIn(t, "nobody@nebula.test", testPassword)

	if wrongPasswordRecorder.Code != http.StatusUnauthorized || unknownEmailRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("got %d and %d, want 401 for both", wrongPasswordRecorder.Code, unknownEmailRecorder.Code)
	}
	if wrongPassword.Error.Message != unknownEmail.Error.Message {
		t.Fatal("wrong password and unknown email must be indistinguishable")
	}
}

func TestLoginEnforcesActivationAndBans(t *testing.T) {
	harness := newLoginTestHarness(t)
	harness.createUser(t, "inactive@nebula.test", false, users.StatusUnverified)
	harness.createUser(t, "banned@nebula.test", true, users.StatusBanned)
	harness.createUser(t, "frozen@nebula.test", true, users.StatusFrozen)

	inactiveRecorder, inactive := harness.logIn(t, "inactive@nebula.test", testPassword)
	bannedRecorder, banned := harness.logIn(t, "banned@nebula.test", testPassword)
	frozenRecorder, _ := harness.logIn(t, "frozen@nebula.test", testPassword)

	if inactiveRecorder.Code != http.StatusForbidden || inactive.Error.Code != "ACCOUNT_NOT_ACTIVATED" {
		t.Fatalf("inactive: got %d %s", inactiveRecorder.Code, inactive.Error.Code)
	}
	if bannedRecorder.Code != http.StatusForbidden || banned.Error.Code != "FORBIDDEN" {
		t.Fatalf("banned: got %d %s", bannedRecorder.Code, banned.Error.Code)
	}
	if frozenRecorder.Code != http.StatusOK {
		t.Fatalf("frozen accounts may log in, got %d", frozenRecorder.Code)
	}
}

func TestRefreshRotatesTheCookieAndRejectsTheOldOne(t *testing.T) {
	harness := newLoginTestHarness(t)
	userID := harness.createUser(t, "pilot@nebula.test", true, users.StatusActive)
	loginRecorder, _ := harness.logIn(t, "pilot@nebula.test", testPassword)
	originalCookie := refreshCookieFrom(t, loginRecorder)

	withCookie := func(refreshCookie *http.Cookie) func(*http.Request) {
		return func(request *http.Request) { request.AddCookie(refreshCookie) }
	}

	refreshRecorder, refreshed := harness.send(t, http.MethodPost, "/api/v1/auth/refresh", nil, withCookie(originalCookie))
	if refreshRecorder.Code != http.StatusOK {
		t.Fatalf("refresh failed with %d %+v", refreshRecorder.Code, refreshed.Error)
	}
	if verifiedUserID, err := harness.accessTokens.Verify(refreshed.Data.AccessToken); err != nil || verifiedUserID != userID {
		t.Fatal("refreshed access token does not verify")
	}
	rotatedCookie := refreshCookieFrom(t, refreshRecorder)
	if rotatedCookie.Value == originalCookie.Value {
		t.Fatal("refresh must rotate the refresh token")
	}

	reusedRecorder, reused := harness.send(t, http.MethodPost, "/api/v1/auth/refresh", nil, withCookie(originalCookie))
	if reusedRecorder.Code != http.StatusUnauthorized || reused.Error.Code != "UNAUTHORIZED" {
		t.Fatalf("old refresh token: got %d %s, want 401", reusedRecorder.Code, reused.Error.Code)
	}
	if clearedCookie := refreshCookieFrom(t, reusedRecorder); clearedCookie.MaxAge >= 0 {
		t.Fatal("a rejected refresh must clear the cookie")
	}

	missingRecorder, _ := harness.send(t, http.MethodPost, "/api/v1/auth/refresh", nil, nil)
	if missingRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("refresh without cookie: got %d", missingRecorder.Code)
	}
}

func TestMeRequiresAValidAccessToken(t *testing.T) {
	harness := newLoginTestHarness(t)
	harness.createUser(t, "pilot@nebula.test", true, users.StatusActive)
	_, loggedIn := harness.logIn(t, "pilot@nebula.test", testPassword)

	authorizedRecorder := httptest.NewRecorder()
	authorizedRequest := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	authorizedRequest.Header.Set("Authorization", "Bearer "+loggedIn.Data.AccessToken)
	harness.router.ServeHTTP(authorizedRecorder, authorizedRequest)

	var profile struct {
		Data users.Profile `json:"data"`
	}
	_ = json.Unmarshal(authorizedRecorder.Body.Bytes(), &profile)
	if authorizedRecorder.Code != http.StatusOK || profile.Data.Email != "pilot@nebula.test" {
		t.Fatalf("got %d %s", authorizedRecorder.Code, authorizedRecorder.Body.String())
	}

	for _, authorizationValue := range []string{"", "Bearer not-a-token", "Basic abc"} {
		unauthorizedRecorder := httptest.NewRecorder()
		unauthorizedRequest := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
		if authorizationValue != "" {
			unauthorizedRequest.Header.Set("Authorization", authorizationValue)
		}
		harness.router.ServeHTTP(unauthorizedRecorder, unauthorizedRequest)
		if unauthorizedRecorder.Code != http.StatusUnauthorized {
			t.Fatalf("authorization %q: got %d, want 401", authorizationValue, unauthorizedRecorder.Code)
		}
	}
}
