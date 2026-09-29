package passwordreset_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/loginlockout"
	"nebula-exchange/backend/internal/auth/passwordhash"
	"nebula-exchange/backend/internal/auth/passwordreset"
	"nebula-exchange/backend/internal/auth/session"
	"nebula-exchange/backend/internal/notify/email"
	"nebula-exchange/backend/internal/notify/email/outbox"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/platform/ratelimit"
	"nebula-exchange/backend/internal/platform/redisclient/redistest"
	"nebula-exchange/backend/internal/users"
)

const originalPassword = "Mining4Crystal!Moon"

var (
	fastHashParameters = passwordhash.Parameters{MemoryInKibibytes: 1024, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}
	resetTokenPattern  = regexp.MustCompile(`/reset\?token=([A-Za-z0-9_%-]+)`)
)

type resetTestHarness struct {
	pool          *pgxpool.Pool
	router        http.Handler
	repository    *users.Repository
	refreshTokens *session.RefreshTokens
	lockout       *loginlockout.Guard
	currentTime   atomic.Pointer[time.Time]
	clientIP      atomic.Int32
}

func (harness *resetTestHarness) now() time.Time { return *harness.currentTime.Load() }

func (harness *resetTestHarness) advanceClock(duration time.Duration) {
	advancedTime := harness.now().Add(duration)
	harness.currentTime.Store(&advancedTime)
}

func newResetTestHarness(t *testing.T) *resetTestHarness {
	t.Helper()
	harness := &resetTestHarness{pool: databasetest.NewPool(t), repository: users.NewRepository()}
	startTime := time.Now().UTC()
	harness.currentTime.Store(&startTime)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	redisClient := redistest.NewClient(t)
	keyPrefix := "test:" + uuid.NewString() + ":"
	templates, err := email.NewTemplateRenderer()
	if err != nil {
		t.Fatalf("load templates: %v", err)
	}

	harness.refreshTokens = session.NewRefreshTokens(session.DefaultRefreshTokenLifetime, harness.now)
	harness.lockout = loginlockout.NewGuard(redisClient, keyPrefix+"lockout:", loginlockout.DefaultPolicy, testLogger)
	service := passwordreset.NewService(passwordreset.Dependencies{
		Pool:            harness.pool,
		Users:           harness.repository,
		PasswordHasher:  passwordhash.NewHasher(passwordhash.HasherOptions{Parameters: fastHashParameters}),
		RefreshTokens:   harness.refreshTokens,
		LoginLockout:    harness.lockout,
		EmailTemplates:  templates,
		EmailQueue:      outbox.NewQueue(),
		FrontendBaseURL: "http://localhost:3000/",
		Now:             harness.now,
	})
	rateLimits := ratelimit.NewMiddlewareFactory(ratelimit.NewSlidingWindowLimiter(redisClient, keyPrefix+"ratelimit:", harness.now), testLogger)
	harness.router = httpserver.NewRouter(httpserver.RouterOptions{Logger: testLogger}, passwordreset.NewHandler(service, rateLimits))
	return harness
}

func (harness *resetTestHarness) createUser(t *testing.T, emailAddress string, isActive bool) uuid.UUID {
	t.Helper()
	passwordHash, _ := passwordhash.Hash(originalPassword, fastHashParameters)
	createdUser, err := harness.repository.Create(context.Background(), harness.pool, users.NewUser{
		Email:           emailAddress,
		Username:        "pilot_" + uuid.NewString()[:8],
		PasswordHash:    passwordHash,
		TermsAcceptedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if isActive {
		if _, err := harness.pool.Exec(context.Background(), "UPDATE users SET is_active = true, activated_at = now() WHERE id = $1", createdUser.ID); err != nil {
			t.Fatalf("activate user: %v", err)
		}
	}
	return createdUser.ID
}

func (harness *resetTestHarness) send(t *testing.T, method, path string, requestBody any) *httptest.ResponseRecorder {
	t.Helper()
	var encodedBody []byte
	if requestBody != nil {
		encodedBody, _ = json.Marshal(requestBody)
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(encodedBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Nebula-Client", "web")
	request.RemoteAddr = "203.0.113." + string(rune('1'+harness.clientIP.Add(1)%8)) + ":40000"
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)
	return recorder
}

func (harness *resetTestHarness) latestResetToken(t *testing.T) string {
	t.Helper()
	var textBody string
	err := harness.pool.QueryRow(context.Background(),
		"SELECT text_body FROM email_outbox WHERE template = 'password_reset' ORDER BY created_at DESC LIMIT 1").Scan(&textBody)
	if err != nil {
		t.Fatalf("no queued reset email: %v", err)
	}
	tokenMatch := resetTokenPattern.FindStringSubmatch(textBody)
	if tokenMatch == nil {
		t.Fatalf("no reset link in %q", textBody)
	}
	plaintextToken, _ := url.QueryUnescape(tokenMatch[1])
	return plaintextToken
}

func (harness *resetTestHarness) queuedResetEmailCount(t *testing.T) int {
	t.Helper()
	var emailCount int
	_ = harness.pool.QueryRow(context.Background(), "SELECT count(*) FROM email_outbox WHERE template = 'password_reset'").Scan(&emailCount)
	return emailCount
}

func (harness *resetTestHarness) storedPasswordMatches(t *testing.T, userID uuid.UUID, candidatePassword string) bool {
	t.Helper()
	var passwordHash string
	_ = harness.pool.QueryRow(context.Background(), "SELECT password_hash FROM users WHERE id = $1", userID).Scan(&passwordHash)
	isMatch, _ := passwordhash.Verify(candidatePassword, passwordHash)
	return isMatch
}

func TestResetRequestQueuesALinkOnlyForActiveAccounts(t *testing.T) {
	harness := newResetTestHarness(t)
	harness.createUser(t, "pilot@nebula.test", true)
	harness.createUser(t, "inactive@nebula.test", false)

	for _, emailAddress := range []string{" PILOT@nebula.test", "inactive@nebula.test", "nobody@nebula.test"} {
		if recorder := harness.send(t, http.MethodPost, "/api/v1/auth/password-reset-requests", map[string]string{"email": emailAddress}); recorder.Code != http.StatusAccepted {
			t.Fatalf("%s: got %d", emailAddress, recorder.Code)
		}
	}

	if queuedCount := harness.queuedResetEmailCount(t); queuedCount != 1 {
		t.Fatalf("only the active account gets an email, got %d", queuedCount)
	}
	previewRecorder := harness.send(t, http.MethodGet, "/api/v1/auth/password-resets/"+harness.latestResetToken(t), nil)
	if previewRecorder.Code != http.StatusOK || !bytes.Contains(previewRecorder.Body.Bytes(), []byte("p•••t@nebula.test")) {
		t.Fatalf("preview: got %d %s", previewRecorder.Code, previewRecorder.Body.String())
	}
}

func TestResetChangesPasswordRevokesSessionsAndClearsLockout(t *testing.T) {
	harness := newResetTestHarness(t)
	userID := harness.createUser(t, "pilot@nebula.test", true)
	issuedRefreshToken, _ := harness.refreshTokens.Issue(context.Background(), harness.pool, userID, session.ClientMetadata{})
	for failedAttempt := 0; failedAttempt < 5; failedAttempt++ {
		_ = harness.lockout.RecordFailure(context.Background(), "pilot@nebula.test")
	}
	harness.send(t, http.MethodPost, "/api/v1/auth/password-reset-requests", map[string]string{"email": "pilot@nebula.test"})
	resetToken := harness.latestResetToken(t)

	recorder := harness.send(t, http.MethodPost, "/api/v1/auth/password-resets", map[string]string{"token": resetToken, "password": "Brand4New!Galaxy"})

	if recorder.Code != http.StatusOK {
		t.Fatalf("reset: got %d %s", recorder.Code, recorder.Body.String())
	}
	if !harness.storedPasswordMatches(t, userID, "Brand4New!Galaxy") || harness.storedPasswordMatches(t, userID, originalPassword) {
		t.Fatal("the stored password must be the new one")
	}
	var revokedAt *time.Time
	_ = harness.pool.QueryRow(context.Background(), "SELECT revoked_at FROM refresh_tokens WHERE id = $1", issuedRefreshToken.ID).Scan(&revokedAt)
	if revokedAt == nil {
		t.Fatal("existing sessions must be revoked")
	}
	if err := harness.lockout.EnsureNotLocked(context.Background(), "pilot@nebula.test"); err != nil {
		t.Fatalf("the login lockout must be cleared, got %v", err)
	}

	if reusedRecorder := harness.send(t, http.MethodPost, "/api/v1/auth/password-resets", map[string]string{"token": resetToken, "password": "Another4New!Pass"}); reusedRecorder.Code != http.StatusNotFound {
		t.Fatalf("a used link must not work twice, got %d", reusedRecorder.Code)
	}
}

func TestNewRequestExpiresOlderLinksAndLinksExpireAfterThirtyMinutes(t *testing.T) {
	harness := newResetTestHarness(t)
	harness.createUser(t, "pilot@nebula.test", true)
	harness.send(t, http.MethodPost, "/api/v1/auth/password-reset-requests", map[string]string{"email": "pilot@nebula.test"})
	olderToken := harness.latestResetToken(t)
	harness.advanceClock(2 * time.Minute)
	harness.send(t, http.MethodPost, "/api/v1/auth/password-reset-requests", map[string]string{"email": "pilot@nebula.test"})
	newerToken := harness.latestResetToken(t)

	if recorder := harness.send(t, http.MethodGet, "/api/v1/auth/password-resets/"+olderToken, nil); recorder.Code != http.StatusNotFound {
		t.Fatalf("the older link must stop working, got %d", recorder.Code)
	}
	if recorder := harness.send(t, http.MethodGet, "/api/v1/auth/password-resets/"+newerToken, nil); recorder.Code != http.StatusOK {
		t.Fatalf("the newer link must work, got %d", recorder.Code)
	}
	harness.advanceClock(31 * time.Minute)
	if recorder := harness.send(t, http.MethodGet, "/api/v1/auth/password-resets/"+newerToken, nil); recorder.Code != http.StatusNotFound {
		t.Fatalf("a link older than 30 minutes must expire, got %d", recorder.Code)
	}
}

func TestResetValidatesThePassword(t *testing.T) {
	harness := newResetTestHarness(t)
	harness.createUser(t, "pilot@nebula.test", true)
	harness.send(t, http.MethodPost, "/api/v1/auth/password-reset-requests", map[string]string{"email": "pilot@nebula.test"})

	recorder := harness.send(t, http.MethodPost, "/api/v1/auth/password-resets", map[string]string{"token": harness.latestResetToken(t), "password": "short"})

	if recorder.Code != http.StatusUnprocessableEntity || !bytes.Contains(recorder.Body.Bytes(), []byte(`"password"`)) {
		t.Fatalf("got %d %s", recorder.Code, recorder.Body.String())
	}
}
