package identity_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity/handler"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity/service"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/config"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database/databasetest"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/email/outbox"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/idempotency"
)

const (
	playerEmail    = "boy@streetborn.test"
	playerUsername = "the_boy"
	playerPassword = "streets-are-cold"
)

type newPlayerProgress struct{}

func (newPlayerProgress) PlayerProgressOf(context.Context, uuid.UUID) (service.PlayerProgress, error) {
	return service.PlayerProgress{FighterLevel: 1}, nil
}

var activationTokenPattern = regexp.MustCompile(`token=([A-Za-z0-9_-]{43})`)

type identityHarness struct {
	t             *testing.T
	router        *gin.Engine
	database      *gorm.DB
	refreshCookie *http.Cookie
}

func newIdentityHarness(t *testing.T) *identityHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)
	testDatabase := databasetest.Open(t)
	quietLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	identityModule, err := identity.NewModule(context.Background(), identity.ModuleDependencies{
		Database:        testDatabase,
		EmailEnqueuer:   outbox.NewEnqueuer(outbox.NewGormRepository(testDatabase), time.Now),
		Session:         config.SessionConfig{JWTSecret: strings.Repeat("s", 40)},
		FrontendBaseURL: "http://localhost:3000",
		PlayerProgress:  newPlayerProgress{},
		Logger:          quietLogger,
		Now:             time.Now,
	})
	if err != nil {
		t.Fatalf("build identity module: %v", err)
	}
	router, err := httpserver.NewRouter(httpserver.RouterOptions{
		Logger:        quietLogger,
		AccessTokens:  identityModule.AccessTokens,
		Idempotency:   idempotency.NewGormStore(testDatabase),
		NonReplayable: identityModule.NonReplayableRoutes(),
	}, identityModule.RouteRegistrars()...)
	if err != nil {
		t.Fatalf("build router: %v", err)
	}
	return &identityHarness{t: t, router: router, database: testDatabase}
}

type apiResponse struct {
	status  int
	body    map[string]any
	cookies []*http.Cookie
}

func (response apiResponse) data() map[string]any {
	data, _ := response.body["data"].(map[string]any)
	return data
}

func (response apiResponse) errorCode() string {
	errorBody, _ := response.body["error"].(map[string]any)
	code, _ := errorBody["code"].(string)
	return code
}

func (harness *identityHarness) send(method, path string, body any, accessToken string) apiResponse {
	harness.t.Helper()
	var requestBody io.Reader
	if body != nil {
		encodedBody, _ := json.Marshal(body)
		requestBody = bytes.NewReader(encodedBody)
	}
	testRequest := httptest.NewRequest(method, "/api/v1"+path, requestBody)
	testRequest.Header.Set("Content-Type", "application/json")
	testRequest.Header.Set("X-Nebula-Client", "web")
	if slices.Contains(handler.NonReplayableRoutes, path) {
		testRequest.Header.Set("Idempotency-Key", "same-key-for-every-session-call")
	}
	if accessToken != "" {
		testRequest.Header.Set("Authorization", "Bearer "+accessToken)
	}
	if harness.refreshCookie != nil {
		testRequest.AddCookie(harness.refreshCookie)
	}
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, testRequest)

	var decodedBody map[string]any
	_ = json.Unmarshal(recorder.Body.Bytes(), &decodedBody)
	return apiResponse{status: recorder.Code, body: decodedBody, cookies: recorder.Result().Cookies()}
}

func (harness *identityHarness) requireStatus(response apiResponse, expectedStatus int) {
	harness.t.Helper()
	if response.status != expectedStatus {
		harness.t.Fatalf("got status %d, want %d, body %v", response.status, expectedStatus, response.body)
	}
}

func (harness *identityHarness) signUp(emailAddress, username string) apiResponse {
	return harness.send(http.MethodPost, "/auth/signup", map[string]any{
		"email": emailAddress, "username": username, "password": playerPassword, "accepts_terms": true,
	}, "")
}

func (harness *identityHarness) activationTokenFromOutbox(emailAddress string) string {
	harness.t.Helper()
	var queuedEmail outbox.OutboxEmail
	if err := harness.database.Where(map[string]any{"recipient_email": emailAddress}).Take(&queuedEmail).Error; err != nil {
		harness.t.Fatalf("find activation email: %v", err)
	}
	tokenMatch := activationTokenPattern.FindStringSubmatch(queuedEmail.TextBody)
	if tokenMatch == nil {
		harness.t.Fatalf("no activation link in email %q", queuedEmail.TextBody)
	}
	return tokenMatch[1]
}

func (harness *identityHarness) logIn(emailAddress, password string) apiResponse {
	loginResponse := harness.send(http.MethodPost, "/auth/login", map[string]any{"email": emailAddress, "password": password}, "")
	harness.keepRefreshCookie(loginResponse)
	return loginResponse
}

func (harness *identityHarness) keepRefreshCookie(response apiResponse) {
	for _, cookie := range response.cookies {
		if cookie.Name == "street_born_refresh_token" && cookie.Value != "" {
			harness.refreshCookie = cookie
		}
	}
}

func TestPlayerSignsUpActivatesLogsInAndReadsProfile(t *testing.T) {
	harness := newIdentityHarness(t)

	signupResponse := harness.signUp(" Boy@StreetBorn.test ", playerUsername)
	harness.requireStatus(signupResponse, http.StatusCreated)
	if signupResponse.data()["status"] != "UNVERIFIED" || signupResponse.data()["email"] != playerEmail {
		t.Fatalf("unexpected signup data %v", signupResponse.data())
	}

	harness.requireStatus(harness.logIn(playerEmail, playerPassword), http.StatusForbidden)
	if code := harness.logIn(playerEmail, playerPassword).errorCode(); code != "ACCOUNT_NOT_ACTIVATED" {
		t.Fatalf("got %s, want ACCOUNT_NOT_ACTIVATED", code)
	}

	activationToken := harness.activationTokenFromOutbox(playerEmail)
	previewResponse := harness.send(http.MethodGet, "/auth/activations/"+activationToken, nil, "")
	harness.requireStatus(previewResponse, http.StatusOK)
	if previewResponse.data()["is_activated"] != false || previewResponse.data()["username"] != playerUsername {
		t.Fatalf("unexpected preview %v", previewResponse.data())
	}

	activationResponse := harness.send(http.MethodPost, "/auth/activations", map[string]any{"token": activationToken}, "")
	harness.requireStatus(activationResponse, http.StatusOK)
	if activationResponse.data()["status"] != "ACTIVE" {
		t.Fatalf("unexpected activation %v", activationResponse.data())
	}
	repeatedActivation := harness.send(http.MethodPost, "/auth/activations", map[string]any{"token": activationToken}, "")
	harness.requireStatus(repeatedActivation, http.StatusOK)
	alreadyActivatedPreview := harness.send(http.MethodGet, "/auth/activations/"+activationToken, nil, "")
	if alreadyActivatedPreview.data()["is_activated"] != true {
		t.Fatalf("expected the preview to say the account is activated, got %v", alreadyActivatedPreview.data())
	}

	harness.requireStatus(harness.logIn(playerEmail, "wrong-password-123"), http.StatusUnauthorized)
	loginResponse := harness.logIn("BOY@streetborn.test", playerPassword)
	harness.requireStatus(loginResponse, http.StatusOK)
	accessToken, _ := loginResponse.data()["access_token"].(string)
	if accessToken == "" || harness.refreshCookie == nil || !harness.refreshCookie.HttpOnly {
		t.Fatalf("expected an access token and an httpOnly refresh cookie, got %v", loginResponse.data())
	}

	harness.requireStatus(harness.send(http.MethodGet, "/me", nil, ""), http.StatusUnauthorized)
	profileResponse := harness.send(http.MethodGet, "/me", nil, accessToken)
	harness.requireStatus(profileResponse, http.StatusOK)
	expectedProfile := map[string]any{"name": playerUsername, "email": playerEmail, "current_level": float64(1), "story_level": nil, "total_win": float64(0), "total_lose": float64(0)}
	if !maps.Equal(profileResponse.data(), expectedProfile) {
		t.Fatalf("got profile %v, want exactly %v", profileResponse.data(), expectedProfile)
	}
}

func TestSignupRejectsTakenIdentifiersAndInvalidFields(t *testing.T) {
	harness := newIdentityHarness(t)
	harness.requireStatus(harness.signUp(playerEmail, playerUsername), http.StatusCreated)

	takenResponse := harness.signUp("BOY@streetborn.test", "THE_BOY")
	harness.requireStatus(takenResponse, http.StatusUnprocessableEntity)
	fieldErrors, _ := takenResponse.body["error"].(map[string]any)["details"].(map[string]any)
	if fieldErrors["email"] == nil || fieldErrors["username"] == nil {
		t.Fatalf("expected both fields to be reported, got %v", fieldErrors)
	}

	invalidResponse := harness.send(http.MethodPost, "/auth/signup", map[string]any{
		"email": "fighter@streetborn.test", "username": "x", "password": "short", "accepts_terms": false,
	}, "")
	harness.requireStatus(invalidResponse, http.StatusUnprocessableEntity)

	harness.requireStatus(harness.send(http.MethodPost, "/auth/signup", nil, ""), http.StatusBadRequest)
}

func TestInvalidOrExpiredActivationLinksAreRejected(t *testing.T) {
	harness := newIdentityHarness(t)
	harness.requireStatus(harness.signUp(playerEmail, playerUsername), http.StatusCreated)
	activationToken := harness.activationTokenFromOutbox(playerEmail)

	harness.requireStatus(harness.send(http.MethodGet, "/auth/activations/not-a-token", nil, ""), http.StatusNotFound)
	harness.requireStatus(harness.send(http.MethodGet, "/auth/activations/"+strings.Repeat("a", 43), nil, ""), http.StatusNotFound)

	expiredAt := time.Now().Add(-time.Hour)
	if err := harness.database.Model(&models.ActivationToken{}).Where("1 = 1").Updates(map[string]any{"created_at": expiredAt.Add(-time.Hour), "expires_at": expiredAt}).Error; err != nil {
		t.Fatalf("expire token: %v", err)
	}
	harness.requireStatus(harness.send(http.MethodGet, "/auth/activations/"+activationToken, nil, ""), http.StatusNotFound)
	harness.requireStatus(harness.send(http.MethodPost, "/auth/activations", map[string]any{"token": activationToken}, ""), http.StatusNotFound)
}

func TestRefreshRotatesTheCookieAndDetectsReuse(t *testing.T) {
	harness := newIdentityHarness(t)
	harness.requireStatus(harness.signUp(playerEmail, playerUsername), http.StatusCreated)
	harness.requireStatus(harness.send(http.MethodPost, "/auth/activations", map[string]any{"token": harness.activationTokenFromOutbox(playerEmail)}, ""), http.StatusOK)
	harness.requireStatus(harness.logIn(playerEmail, playerPassword), http.StatusOK)
	firstCookie := harness.refreshCookie

	refreshResponse := harness.send(http.MethodPost, "/auth/refresh", nil, "")
	harness.requireStatus(refreshResponse, http.StatusOK)
	harness.keepRefreshCookie(refreshResponse)
	if harness.refreshCookie.Value == firstCookie.Value {
		t.Fatal("refresh must rotate the refresh token")
	}
	secondCookie := harness.refreshCookie

	if err := harness.database.Model(&models.RefreshSession{}).Where("revoked_at IS NOT NULL").Update("revoked_at", time.Now().Add(-time.Minute)).Error; err != nil {
		t.Fatalf("age the revoked session: %v", err)
	}
	harness.refreshCookie = firstCookie
	harness.requireStatus(harness.send(http.MethodPost, "/auth/refresh", nil, ""), http.StatusUnauthorized)

	harness.refreshCookie = secondCookie
	harness.requireStatus(harness.send(http.MethodPost, "/auth/refresh", nil, ""), http.StatusUnauthorized)
}

func TestLogoutEndsTheSessionAndBannedPlayersCannotLogIn(t *testing.T) {
	harness := newIdentityHarness(t)
	harness.requireStatus(harness.signUp(playerEmail, playerUsername), http.StatusCreated)
	harness.requireStatus(harness.send(http.MethodPost, "/auth/activations", map[string]any{"token": harness.activationTokenFromOutbox(playerEmail)}, ""), http.StatusOK)
	harness.requireStatus(harness.logIn(playerEmail, playerPassword), http.StatusOK)

	logoutResponse := harness.send(http.MethodPost, "/auth/logout", nil, "")
	harness.requireStatus(logoutResponse, http.StatusOK)
	harness.requireStatus(harness.send(http.MethodPost, "/auth/refresh", nil, ""), http.StatusUnauthorized)

	if err := harness.database.Model(&models.User{}).Where(map[string]any{"email": playerEmail}).Update("status", models.UserStatusBanned).Error; err != nil {
		t.Fatalf("ban player: %v", err)
	}
	bannedResponse := harness.logIn(playerEmail, playerPassword)
	harness.requireStatus(bannedResponse, http.StatusForbidden)
	if bannedResponse.errorCode() != "FORBIDDEN" {
		t.Fatalf("got %s, want FORBIDDEN", bannedResponse.errorCode())
	}
}
