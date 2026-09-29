package activation_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/activation"
	"nebula-exchange/backend/internal/notify/email"
	"nebula-exchange/backend/internal/notify/email/outbox"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/platform/ratelimit"
	"nebula-exchange/backend/internal/platform/redisclient/redistest"
	"nebula-exchange/backend/internal/users"
)

var activationLinkTokenPattern = regexp.MustCompile(`/activate\?token=([A-Za-z0-9_%-]+)`)

type resendTestHarness struct {
	pool       *pgxpool.Pool
	router     http.Handler
	repository *users.Repository
	issuer     *activation.Issuer
	clientIP   string
}

func newResendTestHarness(t *testing.T) *resendTestHarness {
	t.Helper()
	pool := databasetest.NewPool(t)
	templateRenderer, err := email.NewTemplateRenderer()
	if err != nil {
		t.Fatalf("load templates: %v", err)
	}
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	repository := users.NewRepository()
	issuer := activation.NewIssuer(activation.DefaultTokenLifetime, time.Now)
	mailer := activation.NewMailer(activation.NewEmailComposer("http://localhost:3000", templateRenderer), outbox.NewQueue())
	rateLimits := ratelimit.NewMiddlewareFactory(
		ratelimit.NewSlidingWindowLimiter(redistest.NewClient(t), "test:ratelimit:"+uuid.NewString()+":", time.Now),
		testLogger,
	)

	router := httpserver.NewRouter(
		httpserver.RouterOptions{Logger: testLogger},
		activation.NewHandler(activation.NewService(pool, repository, time.Now)),
		activation.NewResendHandler(activation.NewResender(pool, repository, issuer, mailer, time.Now), rateLimits),
	)
	return &resendTestHarness{pool: pool, router: router, repository: repository, issuer: issuer, clientIP: "203.0.113.10"}
}

func (harness *resendTestHarness) createUser(t *testing.T, emailAddress string, isActive bool) uuid.UUID {
	t.Helper()
	createdUser, err := harness.repository.Create(context.Background(), harness.pool, users.NewUser{
		Email:           emailAddress,
		Username:        "pilot_" + uuid.NewString()[:8],
		PasswordHash:    "unused-in-this-test",
		TermsAcceptedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if isActive {
		_, err = harness.pool.Exec(context.Background(), "UPDATE users SET is_active = true, activated_at = now() WHERE id = $1", createdUser.ID)
		if err != nil {
			t.Fatalf("activate user: %v", err)
		}
	}
	return createdUser.ID
}

func (harness *resendTestHarness) requestResend(t *testing.T, emailAddress string) *httptest.ResponseRecorder {
	t.Helper()
	encodedBody, _ := json.Marshal(map[string]string{"email": emailAddress})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/activation-emails", bytes.NewReader(encodedBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Nebula-Client", "web")
	request.RemoteAddr = harness.clientIP + ":40000"
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)
	return recorder
}

func (harness *resendTestHarness) queuedEmailBodies(t *testing.T) []string {
	t.Helper()
	rows, err := harness.pool.Query(context.Background(), "SELECT text_body FROM email_outbox ORDER BY created_at")
	if err != nil {
		t.Fatalf("query outbox: %v", err)
	}
	defer rows.Close()
	var textBodies []string
	for rows.Next() {
		var textBody string
		_ = rows.Scan(&textBody)
		textBodies = append(textBodies, textBody)
	}
	return textBodies
}

func (harness *resendTestHarness) previewStatus(t *testing.T, plaintextToken string) int {
	t.Helper()
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/auth/activations/"+plaintextToken, nil))
	return recorder.Code
}

func tokenFromEmailBody(t *testing.T, textBody string) string {
	t.Helper()
	tokenMatch := activationLinkTokenPattern.FindStringSubmatch(textBody)
	if tokenMatch == nil {
		t.Fatalf("no activation link in %q", textBody)
	}
	plaintextToken, _ := url.QueryUnescape(tokenMatch[1])
	return plaintextToken
}

func TestResendQueuesANewLinkAndExpiresTheOldOne(t *testing.T) {
	harness := newResendTestHarness(t)
	userID := harness.createUser(t, "pilot@nebula.test", false)
	originalToken, err := harness.issuer.Issue(context.Background(), harness.pool, userID)
	if err != nil {
		t.Fatalf("issue original token: %v", err)
	}

	recorder := harness.requestResend(t, "  PILOT@nebula.test ")

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("got status %d %s", recorder.Code, recorder.Body.String())
	}
	queuedBodies := harness.queuedEmailBodies(t)
	if len(queuedBodies) != 1 {
		t.Fatalf("expected one queued email, got %d", len(queuedBodies))
	}
	newToken := tokenFromEmailBody(t, queuedBodies[0])
	if newToken == originalToken.Plaintext {
		t.Fatal("resend must issue a new token")
	}
	if statusCode := harness.previewStatus(t, newToken); statusCode != http.StatusOK {
		t.Fatalf("new link must work, got %d", statusCode)
	}
	if statusCode := harness.previewStatus(t, originalToken.Plaintext); statusCode != http.StatusNotFound {
		t.Fatalf("old link must stop working, got %d", statusCode)
	}
}

func TestResendAnswersTheSameForUnknownAndActiveAccounts(t *testing.T) {
	harness := newResendTestHarness(t)
	harness.createUser(t, "active@nebula.test", true)

	unknownRecorder := harness.requestResend(t, "nobody@nebula.test")
	activeRecorder := harness.requestResend(t, "active@nebula.test")

	if unknownRecorder.Code != http.StatusAccepted || activeRecorder.Code != http.StatusAccepted {
		t.Fatalf("got %d and %d, want 202 for both", unknownRecorder.Code, activeRecorder.Code)
	}
	if unknownRecorder.Body.String() != activeRecorder.Body.String() {
		t.Fatal("responses must be identical so they reveal nothing")
	}
	if queuedBodies := harness.queuedEmailBodies(t); len(queuedBodies) != 0 {
		t.Fatalf("no email should be queued, got %d", len(queuedBodies))
	}
}

func TestResendIsLimitedPerEmailAndPerIP(t *testing.T) {
	harness := newResendTestHarness(t)
	harness.createUser(t, "pilot@nebula.test", false)

	firstRecorder := harness.requestResend(t, "pilot@nebula.test")
	secondRecorder := harness.requestResend(t, "Pilot@Nebula.test")
	if firstRecorder.Code != http.StatusAccepted || secondRecorder.Code != http.StatusTooManyRequests {
		t.Fatalf("per-email limit: got %d then %d, want 202 then 429", firstRecorder.Code, secondRecorder.Code)
	}
	if secondRecorder.Header().Get("Retry-After") == "" {
		t.Fatal("a limited response must include Retry-After")
	}

	harness.clientIP = "198.51.100.20"
	for requestNumber := 1; requestNumber <= 3; requestNumber++ {
		if recorder := harness.requestResend(t, uuid.NewString()+"@nebula.test"); recorder.Code != http.StatusAccepted {
			t.Fatalf("request %d from a fresh IP: got %d", requestNumber, recorder.Code)
		}
	}
	if recorder := harness.requestResend(t, uuid.NewString()+"@nebula.test"); recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("fourth request from the same IP: got %d, want 429", recorder.Code)
	}
}
