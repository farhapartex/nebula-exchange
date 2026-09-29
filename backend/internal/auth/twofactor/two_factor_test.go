package twofactor_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp/totp"

	"nebula-exchange/backend/internal/auth/twofactor"
	"nebula-exchange/backend/internal/notify/email"
	"nebula-exchange/backend/internal/notify/email/outbox"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/secretbox"
	"nebula-exchange/backend/internal/users"
)

type twoFactorHarness struct {
	pool        *pgxpool.Pool
	service     *twofactor.Service
	currentTime atomic.Pointer[time.Time]
}

func (harness *twoFactorHarness) now() time.Time { return *harness.currentTime.Load() }

func (harness *twoFactorHarness) advanceClock(duration time.Duration) {
	advancedTime := harness.now().Add(duration)
	harness.currentTime.Store(&advancedTime)
}

func newTwoFactorHarness(t *testing.T) *twoFactorHarness {
	t.Helper()
	harness := &twoFactorHarness{pool: databasetest.NewPool(t)}
	startTime := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	harness.currentTime.Store(&startTime)

	encryptionKey := make([]byte, 32)
	_, _ = rand.Read(encryptionKey)
	secretBox, err := secretbox.NewFromBase64Key(base64.StdEncoding.EncodeToString(encryptionKey))
	if err != nil {
		t.Fatalf("create secret box: %v", err)
	}
	templates, err := email.NewTemplateRenderer()
	if err != nil {
		t.Fatalf("load templates: %v", err)
	}
	harness.service = twofactor.NewService(twofactor.Dependencies{
		Pool:           harness.pool,
		SecretBox:      secretBox,
		EmailTemplates: templates,
		EmailQueue:     outbox.NewQueue(),
		Now:            harness.now,
	})
	return harness
}

func (harness *twoFactorHarness) createUser(t *testing.T) uuid.UUID {
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
	return createdUser.ID
}

func (harness *twoFactorHarness) codeFor(t *testing.T, secret string) string {
	t.Helper()
	code, err := totp.GenerateCode(secret, harness.now())
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}
	return code
}

func requireIncorrectCode(t *testing.T, err error) {
	t.Helper()
	var apiError *apierror.Error
	if !errors.As(err, &apiError) || apiError.Code != apierror.CodeValidationFailed {
		t.Fatalf("got %v, want an incorrect code error", err)
	}
}

func TestSetupStoresAnEncryptedSecretAndEnableNeedsAValidCode(t *testing.T) {
	harness := newTwoFactorHarness(t)
	userID := harness.createUser(t)

	setupDetails, err := harness.service.BeginSetup(context.Background(), userID)
	if err != nil {
		t.Fatalf("begin setup: %v", err)
	}
	var storedSecret []byte
	_ = harness.pool.QueryRow(context.Background(), "SELECT totp_secret_enc FROM users WHERE id = $1", userID).Scan(&storedSecret)
	if len(storedSecret) == 0 || string(storedSecret) == setupDetails.Secret {
		t.Fatal("the secret must be stored encrypted")
	}

	requireIncorrectCode(t, harness.service.Enable(context.Background(), userID, "000000"))
	if err := harness.service.Enable(context.Background(), userID, harness.codeFor(t, setupDetails.Secret)); err != nil {
		t.Fatalf("enable: %v", err)
	}

	var queuedEmailCount int
	_ = harness.pool.QueryRow(context.Background(), "SELECT count(*) FROM email_outbox WHERE template = 'two_factor_changed'").Scan(&queuedEmailCount)
	if queuedEmailCount != 1 {
		t.Fatalf("enabling must queue a notification email, got %d", queuedEmailCount)
	}
	if _, err := harness.service.BeginSetup(context.Background(), userID); err == nil {
		t.Fatal("setup must be refused while 2FA is on")
	}
}

func TestCodesCannotBeReusedAndDisableNeedsAFreshCode(t *testing.T) {
	harness := newTwoFactorHarness(t)
	userID := harness.createUser(t)
	setupDetails, _ := harness.service.BeginSetup(context.Background(), userID)
	enableCode := harness.codeFor(t, setupDetails.Secret)
	if err := harness.service.Enable(context.Background(), userID, enableCode); err != nil {
		t.Fatalf("enable: %v", err)
	}

	requireIncorrectCode(t, harness.service.VerifyCode(context.Background(), userID, enableCode))

	harness.advanceClock(30 * time.Second)
	if err := harness.service.VerifyCode(context.Background(), userID, harness.codeFor(t, setupDetails.Secret)); err != nil {
		t.Fatalf("a fresh code must verify: %v", err)
	}

	harness.advanceClock(30 * time.Second)
	if err := harness.service.Disable(context.Background(), userID, harness.codeFor(t, setupDetails.Secret)); err != nil {
		t.Fatalf("disable: %v", err)
	}
	var enabledAt *time.Time
	var storedSecret []byte
	_ = harness.pool.QueryRow(context.Background(), "SELECT totp_enabled_at, totp_secret_enc FROM users WHERE id = $1", userID).Scan(&enabledAt, &storedSecret)
	if enabledAt != nil || storedSecret != nil {
		t.Fatal("disabling must remove the secret")
	}
}

func TestCodesFromTheNeighbouringWindowAreAccepted(t *testing.T) {
	harness := newTwoFactorHarness(t)
	userID := harness.createUser(t)
	setupDetails, _ := harness.service.BeginSetup(context.Background(), userID)
	previousWindowCode, _ := totp.GenerateCode(setupDetails.Secret, harness.now().Add(-30*time.Second))

	if err := harness.service.Enable(context.Background(), userID, previousWindowCode); err != nil {
		t.Fatalf("a code from the previous 30 seconds must be accepted for clock drift: %v", err)
	}
	staleCode, _ := totp.GenerateCode(setupDetails.Secret, harness.now().Add(-2*time.Minute))
	requireIncorrectCode(t, harness.service.VerifyCode(context.Background(), userID, staleCode))
}
