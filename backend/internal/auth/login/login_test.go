package login_test

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

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/accesstoken"
	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/auth/login"
	"nebula-exchange/backend/internal/auth/loginlockout"
	"nebula-exchange/backend/internal/auth/passwordhash"
	"nebula-exchange/backend/internal/auth/session"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/platform/redisclient/redistest"
	"nebula-exchange/backend/internal/users"
)

const testPassword = "Mining4Crystal!Moon"

var fastHashParameters = passwordhash.Parameters{MemoryInKibibytes: 1024, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}

type loginTestHarness struct {
	pool         *pgxpool.Pool
	router       http.Handler
	accessTokens *accesstoken.Manager
	repository   *users.Repository
	currentTime  atomic.Pointer[time.Time]
}

func (harness *loginTestHarness) now() time.Time { return *harness.currentTime.Load() }

func (harness *loginTestHarness) advanceClock(duration time.Duration) {
	advancedTime := harness.now().Add(duration)
	harness.currentTime.Store(&advancedTime)
}

func newLoginTestHarness(t *testing.T) *loginTestHarness {
	t.Helper()
	pool := databasetest.NewPool(t)
	accessTokens, err := accesstoken.NewManager("login-test-secret-with-at-least-32-chars", accesstoken.DefaultLifetime, time.Now)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	repository := users.NewRepository()
	harness := &loginTestHarness{pool: pool, accessTokens: accessTokens, repository: repository}
	startTime := time.Now()
	harness.currentTime.Store(&startTime)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)

	loginService, err := login.NewService(login.Dependencies{
		Pool:          pool,
		Users:         repository,
		AccessTokens:  accessTokens,
		RefreshTokens: session.NewRefreshTokens(session.DefaultRefreshTokenLifetime, harness.now),
		LoginLockout: loginlockout.NewGuard(
			redistest.NewClient(t),
			"test:login-lockout:"+uuid.NewString()+":",
			loginlockout.DefaultPolicy,
			testLogger,
		),
		Logger:              testLogger,
		PasswordHasher:      passwordhash.NewHasher(passwordhash.HasherOptions{Parameters: fastHashParameters}),
		PasswordHashOptions: fastHashParameters,
		Now:                 harness.now,
	})
	if err != nil {
		t.Fatalf("create login service: %v", err)
	}

	harness.router = httpserver.NewRouter(
		httpserver.RouterOptions{Logger: testLogger, IdentifyUser: authentication.IdentifyUser(accessTokens)},
		login.NewHandler(loginService, session.CookieSettings{IsSecure: true}, harness.now),
		users.NewMeHandler(pool, repository),
	)
	return harness
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

func withRefreshCookie(refreshCookie *http.Cookie) func(*http.Request) {
	return func(request *http.Request) { request.AddCookie(refreshCookie) }
}

func (harness *loginTestHarness) refresh(t *testing.T, refreshCookie *http.Cookie) (*httptest.ResponseRecorder, sessionEnvelope) {
	t.Helper()
	return harness.send(t, http.MethodPost, "/api/v1/auth/refresh", nil, withRefreshCookie(refreshCookie))
}

func TestLogoutRevokesTheRefreshTokenAndClearsTheCookie(t *testing.T) {
	harness := newLoginTestHarness(t)
	harness.createUser(t, "pilot@nebula.test", true, users.StatusActive)
	loginRecorder, _ := harness.logIn(t, "pilot@nebula.test", testPassword)
	refreshCookie := refreshCookieFrom(t, loginRecorder)

	logoutRecorder, _ := harness.send(t, http.MethodPost, "/api/v1/auth/logout", nil, withRefreshCookie(refreshCookie))

	if logoutRecorder.Code != http.StatusOK {
		t.Fatalf("logout: got %d %s", logoutRecorder.Code, logoutRecorder.Body.String())
	}
	if clearedCookie := refreshCookieFrom(t, logoutRecorder); clearedCookie.MaxAge >= 0 {
		t.Fatal("logout must clear the refresh cookie")
	}
	if refreshRecorder, _ := harness.refresh(t, refreshCookie); refreshRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("refresh after logout: got %d, want 401", refreshRecorder.Code)
	}

	anonymousRecorder, _ := harness.send(t, http.MethodPost, "/api/v1/auth/logout", nil, nil)
	if anonymousRecorder.Code != http.StatusOK {
		t.Fatalf("logout without a session must still succeed, got %d", anonymousRecorder.Code)
	}
}

func TestReusingAnOldRefreshTokenRevokesEverySession(t *testing.T) {
	harness := newLoginTestHarness(t)
	harness.createUser(t, "pilot@nebula.test", true, users.StatusActive)
	laptopLogin, _ := harness.logIn(t, "pilot@nebula.test", testPassword)
	phoneLogin, _ := harness.logIn(t, "pilot@nebula.test", testPassword)
	stolenLaptopCookie := refreshCookieFrom(t, laptopLogin)

	laptopRefresh, _ := harness.refresh(t, stolenLaptopCookie)
	rotatedLaptopCookie := refreshCookieFrom(t, laptopRefresh)
	harness.advanceClock(session.DefaultConcurrentRefreshGrace + time.Second)

	reuseRecorder, reused := harness.refresh(t, stolenLaptopCookie)
	if reuseRecorder.Code != http.StatusUnauthorized || reused.Error.Code != "UNAUTHORIZED" {
		t.Fatalf("reused token: got %d %s", reuseRecorder.Code, reused.Error.Code)
	}

	if recorder, _ := harness.refresh(t, rotatedLaptopCookie); recorder.Code != http.StatusUnauthorized {
		t.Fatalf("the rotated laptop session must be revoked, got %d", recorder.Code)
	}
	if recorder, _ := harness.refresh(t, refreshCookieFrom(t, phoneLogin)); recorder.Code != http.StatusUnauthorized {
		t.Fatalf("the phone session must be revoked too, got %d", recorder.Code)
	}
}

func TestConcurrentRefreshWithinGraceDoesNotRevokeSessions(t *testing.T) {
	harness := newLoginTestHarness(t)
	harness.createUser(t, "pilot@nebula.test", true, users.StatusActive)
	loginRecorder, _ := harness.logIn(t, "pilot@nebula.test", testPassword)
	sharedCookie := refreshCookieFrom(t, loginRecorder)

	firstTabRefresh, _ := harness.refresh(t, sharedCookie)
	harness.advanceClock(2 * time.Second)
	secondTabRecorder, _ := harness.refresh(t, sharedCookie)

	if secondTabRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("the losing tab gets a plain 401, got %d", secondTabRecorder.Code)
	}
	if recorder, _ := harness.refresh(t, refreshCookieFrom(t, firstTabRefresh)); recorder.Code != http.StatusOK {
		t.Fatalf("the winning session must stay valid after a concurrent refresh, got %d", recorder.Code)
	}
}

type lockedEnvelope struct {
	Error struct {
		Code    string         `json:"code"`
		Details map[string]int `json:"details"`
	} `json:"error"`
}

func (harness *loginTestHarness) failLogins(t *testing.T, emailAddress string, attemptCount int) *httptest.ResponseRecorder {
	t.Helper()
	var lastRecorder *httptest.ResponseRecorder
	for attemptNumber := 0; attemptNumber < attemptCount; attemptNumber++ {
		lastRecorder, _ = harness.logIn(t, emailAddress, "definitely-wrong-password")
	}
	return lastRecorder
}

func decodeLocked(t *testing.T, recorder *httptest.ResponseRecorder) lockedEnvelope {
	t.Helper()
	var lockedBody lockedEnvelope
	_ = json.Unmarshal(recorder.Body.Bytes(), &lockedBody)
	return lockedBody
}

func TestFiveFailedLoginsLockTheEmail(t *testing.T) {
	harness := newLoginTestHarness(t)
	harness.createUser(t, "pilot@nebula.test", true, users.StatusActive)

	if fourthRecorder := harness.failLogins(t, "pilot@nebula.test", 4); fourthRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("fourth failure should still be 401, got %d", fourthRecorder.Code)
	}
	fifthRecorder := harness.failLogins(t, "pilot@nebula.test", 1)
	lockedBody := decodeLocked(t, fifthRecorder)
	if fifthRecorder.Code != http.StatusTooManyRequests || lockedBody.Error.Code != "LOGIN_LOCKED" {
		t.Fatalf("fifth failure should lock, got %d %s", fifthRecorder.Code, lockedBody.Error.Code)
	}
	if retryAfterSeconds := lockedBody.Error.Details["retry_after_seconds"]; retryAfterSeconds < 890 || retryAfterSeconds > 900 {
		t.Fatalf("got retry_after_seconds %d, want about 900", retryAfterSeconds)
	}

	correctPasswordRecorder, _ := harness.logIn(t, "PILOT@nebula.test", testPassword)
	if correctPasswordRecorder.Code != http.StatusTooManyRequests {
		t.Fatalf("even the right password must be refused while locked, got %d", correctPasswordRecorder.Code)
	}

	harness.createUser(t, "other@nebula.test", true, users.StatusActive)
	if otherRecorder, _ := harness.logIn(t, "other@nebula.test", testPassword); otherRecorder.Code != http.StatusOK {
		t.Fatalf("the lock must only affect the locked email, got %d", otherRecorder.Code)
	}
}

func TestUnknownEmailsLockTheSameWay(t *testing.T) {
	harness := newLoginTestHarness(t)

	lockingRecorder := harness.failLogins(t, "nobody@nebula.test", 5)

	if lockingRecorder.Code != http.StatusTooManyRequests || decodeLocked(t, lockingRecorder).Error.Code != "LOGIN_LOCKED" {
		t.Fatalf("unknown emails must lock like real ones, got %d", lockingRecorder.Code)
	}
}

func TestSuccessfulLoginResetsTheFailureCount(t *testing.T) {
	harness := newLoginTestHarness(t)
	harness.createUser(t, "pilot@nebula.test", true, users.StatusActive)

	harness.failLogins(t, "pilot@nebula.test", 4)
	if successRecorder, _ := harness.logIn(t, "pilot@nebula.test", testPassword); successRecorder.Code != http.StatusOK {
		t.Fatalf("login should succeed before the limit, got %d", successRecorder.Code)
	}

	if afterResetRecorder := harness.failLogins(t, "pilot@nebula.test", 4); afterResetRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("the counter should restart after a success, got %d", afterResetRecorder.Code)
	}
}
