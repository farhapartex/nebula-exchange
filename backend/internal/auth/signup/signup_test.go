package signup_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/activation"
	"nebula-exchange/backend/internal/auth/passwordhash"
	"nebula-exchange/backend/internal/auth/signup"
	"nebula-exchange/backend/internal/notify/email"
	"nebula-exchange/backend/internal/notify/email/outbox"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/users"
)

var fixedNow = time.Date(2026, 9, 30, 10, 0, 0, 0, time.UTC)

type signupTestHarness struct {
	pool   *pgxpool.Pool
	router http.Handler
}

func newSignupTestHarness(t *testing.T) *signupTestHarness {
	t.Helper()
	pool := databasetest.NewPool(t)
	templateRenderer, err := email.NewTemplateRenderer()
	if err != nil {
		t.Fatalf("load email templates: %v", err)
	}
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	clock := func() time.Time { return fixedNow }

	signupService := signup.NewService(signup.Dependencies{
		Pool:             pool,
		Users:            users.NewRepository(),
		ActivationIssuer: activation.NewIssuer(activation.DefaultTokenLifetime, clock),
		ActivationEmail:  activation.NewEmailComposer("http://localhost:3000/", templateRenderer),
		EmailQueue:       outbox.NewQueue(),
		PasswordHasher: passwordhash.NewHasher(passwordhash.HasherOptions{
			Parameters: passwordhash.Parameters{MemoryInKibibytes: 1024, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32},
		}),
		Now: clock,
	})

	router := httpserver.NewRouter(httpserver.RouterOptions{Logger: testLogger}, signup.NewHandler(signupService))
	return &signupTestHarness{pool: pool, router: router}
}

type signupResponse struct {
	Data struct {
		ID                      string    `json:"id"`
		Email                   string    `json:"email"`
		Username                string    `json:"username"`
		Status                  string    `json:"status"`
		ActivationLinkExpiresAt time.Time `json:"activation_link_expires_at"`
	} `json:"data"`
	Error struct {
		Code    string            `json:"code"`
		Details map[string]string `json:"details"`
	} `json:"error"`
}

func (harness *signupTestHarness) signUp(t *testing.T, requestBody map[string]any) (int, signupResponse) {
	t.Helper()
	encodedBody, _ := json.Marshal(requestBody)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(encodedBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Nebula-Client", "web")
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)

	var decodedResponse signupResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &decodedResponse); err != nil {
		t.Fatalf("decode response %s: %v", recorder.Body.String(), err)
	}
	return recorder.Code, decodedResponse
}

func validSignupBody() map[string]any {
	return map[string]any{
		"email":         "  Pilot@Nebula.Test ",
		"username":      "pilot_nova",
		"password":      "Mining4Crystal!Moon",
		"accepts_terms": true,
	}
}

func TestSignupCreatesInactiveUserAndSendsActivationLink(t *testing.T) {
	harness := newSignupTestHarness(t)

	statusCode, signedUp := harness.signUp(t, validSignupBody())

	if statusCode != http.StatusCreated {
		t.Fatalf("got status %d, want 201, error %+v", statusCode, signedUp.Error)
	}
	if signedUp.Data.Email != "pilot@nebula.test" || signedUp.Data.Username != "pilot_nova" || signedUp.Data.Status != "UNVERIFIED" {
		t.Fatalf("unexpected account %+v", signedUp.Data)
	}
	if !signedUp.Data.ActivationLinkExpiresAt.Equal(fixedNow.Add(24 * time.Hour)) {
		t.Fatalf("got expiry %v, want 24 hours after signup", signedUp.Data.ActivationLinkExpiresAt)
	}

	var storedPasswordHash string
	var isActive bool
	var activatedAt, termsAcceptedAt *time.Time
	err := harness.pool.QueryRow(context.Background(),
		"SELECT password_hash, is_active, activated_at, terms_accepted_at FROM users WHERE id = $1", signedUp.Data.ID,
	).Scan(&storedPasswordHash, &isActive, &activatedAt, &termsAcceptedAt)
	if err != nil {
		t.Fatalf("load created user: %v", err)
	}
	if isActive || activatedAt != nil {
		t.Fatal("a new account must start inactive")
	}
	if termsAcceptedAt == nil || !termsAcceptedAt.Equal(fixedNow) {
		t.Fatalf("got terms_accepted_at %v, want %v", termsAcceptedAt, fixedNow)
	}
	if isMatch, _ := passwordhash.Verify("Mining4Crystal!Moon", storedPasswordHash); !isMatch {
		t.Fatal("stored password hash does not verify")
	}

	var queuedRecipient, queuedTemplate, queuedStatus, queuedTextBody, queuedHTMLBody string
	err = harness.pool.QueryRow(context.Background(),
		"SELECT recipient_email, template, status, text_body, html_body FROM email_outbox",
	).Scan(&queuedRecipient, &queuedTemplate, &queuedStatus, &queuedTextBody, &queuedHTMLBody)
	if err != nil {
		t.Fatalf("expected one queued activation email: %v", err)
	}
	if queuedRecipient != "pilot@nebula.test" || queuedTemplate != "account_activation" || queuedStatus != "pending" {
		t.Fatalf("unexpected outbox row %s %s %s", queuedRecipient, queuedTemplate, queuedStatus)
	}
	plaintextToken := extractActivationToken(t, queuedTextBody)

	var storedTokenHash []byte
	err = harness.pool.QueryRow(context.Background(),
		"SELECT token_hash FROM account_activation_tokens WHERE user_id = $1", signedUp.Data.ID,
	).Scan(&storedTokenHash)
	if err != nil {
		t.Fatalf("load activation token: %v", err)
	}
	if !bytes.Equal(storedTokenHash, activation.HashToken(plaintextToken)) {
		t.Fatal("stored token hash does not match the emailed token")
	}
	if bytes.Contains(storedTokenHash, []byte(plaintextToken)) {
		t.Fatal("the plaintext token must never be stored")
	}
	if !strings.Contains(queuedHTMLBody, "http://localhost:3000/activate?token=") {
		t.Fatal("html email must contain the activation link")
	}
}

func extractActivationToken(t *testing.T, textBody string) string {
	t.Helper()
	for _, line := range strings.Split(textBody, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "http://localhost:3000/activate?token=") {
			activationLink, err := url.Parse(strings.TrimSpace(line))
			if err != nil {
				t.Fatalf("parse activation link: %v", err)
			}
			return activationLink.Query().Get("token")
		}
	}
	t.Fatalf("no activation link in email body:\n%s", textBody)
	return ""
}

func TestSignupRejectsTakenEmailAndUsernameCaseInsensitively(t *testing.T) {
	harness := newSignupTestHarness(t)
	harness.signUp(t, validSignupBody())

	duplicateBody := validSignupBody()
	duplicateBody["email"] = "PILOT@nebula.test"
	duplicateBody["username"] = "Pilot_Nova"
	statusCode, rejected := harness.signUp(t, duplicateBody)

	if statusCode != http.StatusUnprocessableEntity || rejected.Error.Code != "VALIDATION_FAILED" {
		t.Fatalf("got %d %s, want 422 VALIDATION_FAILED", statusCode, rejected.Error.Code)
	}
	if rejected.Error.Details["email"] != "is already registered" || rejected.Error.Details["username"] != "is already taken" {
		t.Fatalf("unexpected field errors %v", rejected.Error.Details)
	}

	onlyUsernameBody := validSignupBody()
	onlyUsernameBody["email"] = "other@nebula.test"
	_, usernameRejected := harness.signUp(t, onlyUsernameBody)
	if _, hasEmailError := usernameRejected.Error.Details["email"]; hasEmailError || usernameRejected.Error.Details["username"] == "" {
		t.Fatalf("expected only a username error, got %v", usernameRejected.Error.Details)
	}
}

func TestSignupValidatesFields(t *testing.T) {
	harness := newSignupTestHarness(t)

	invalidBodies := map[string]struct {
		changes       map[string]any
		expectedField string
	}{
		"invalid email":      {changes: map[string]any{"email": "not-an-email"}, expectedField: "email"},
		"short password":     {changes: map[string]any{"password": "short"}, expectedField: "password"},
		"username with dash": {changes: map[string]any{"username": "pilot-nova"}, expectedField: "username"},
		"terms not accepted": {changes: map[string]any{"accepts_terms": false}, expectedField: "accepts_terms"},
	}

	for caseName, invalidBody := range invalidBodies {
		t.Run(caseName, func(t *testing.T) {
			requestBody := validSignupBody()
			for fieldName, fieldValue := range invalidBody.changes {
				requestBody[fieldName] = fieldValue
			}
			statusCode, rejected := harness.signUp(t, requestBody)
			if statusCode != http.StatusUnprocessableEntity {
				t.Fatalf("got status %d, want 422", statusCode)
			}
			if rejected.Error.Details[invalidBody.expectedField] == "" {
				t.Fatalf("expected an error on %s, got %v", invalidBody.expectedField, rejected.Error.Details)
			}
		})
	}

	var userCount int
	if err := harness.pool.QueryRow(context.Background(), "SELECT count(*) FROM users").Scan(&userCount); err != nil || userCount != 0 {
		t.Fatalf("invalid signups must not create users, found %d (%v)", userCount, err)
	}
}

func TestRejectedSignupQueuesNoEmail(t *testing.T) {
	harness := newSignupTestHarness(t)
	requestBody := validSignupBody()
	requestBody["accepts_terms"] = false

	harness.signUp(t, requestBody)

	var queuedEmailCount int
	if err := harness.pool.QueryRow(context.Background(), "SELECT count(*) FROM email_outbox").Scan(&queuedEmailCount); err != nil || queuedEmailCount != 0 {
		t.Fatalf("a rejected signup must not queue email, found %d (%v)", queuedEmailCount, err)
	}
}
