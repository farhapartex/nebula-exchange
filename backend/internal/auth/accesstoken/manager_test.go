package accesstoken

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const testSecret = "test-secret-that-is-long-enough-for-hs256"

func newTestManager(t *testing.T, now time.Time) *Manager {
	t.Helper()
	manager, err := NewManager(testSecret, DefaultLifetime, func() time.Time { return now })
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	return manager
}

func TestIssueAndVerifyRoundTrip(t *testing.T) {
	issuedAt := time.Date(2026, 9, 30, 10, 0, 0, 0, time.UTC)
	manager := newTestManager(t, issuedAt)
	userID := uuid.New()

	issuedToken, err := manager.Issue(userID)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	if !issuedToken.ExpiresAt.Equal(issuedAt.Add(15 * time.Minute)) {
		t.Fatalf("got expiry %v", issuedToken.ExpiresAt)
	}

	verifiedUserID, err := manager.Verify(issuedToken.Value)
	if err != nil || verifiedUserID != userID {
		t.Fatalf("got %v, %v, want %v", verifiedUserID, err, userID)
	}
}

func TestVerifyRejectsExpiredTamperedAndForeignTokens(t *testing.T) {
	issuedAt := time.Date(2026, 9, 30, 10, 0, 0, 0, time.UTC)
	issuedToken, _ := newTestManager(t, issuedAt).Issue(uuid.New())

	laterManager := newTestManager(t, issuedAt.Add(16*time.Minute))
	if _, err := laterManager.Verify(issuedToken.Value); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expired token: got %v", err)
	}

	tamperedToken := issuedToken.Value[:len(issuedToken.Value)-2] + "xx"
	if _, err := newTestManager(t, issuedAt).Verify(tamperedToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("tampered token: got %v", err)
	}

	foreignManager, _ := NewManager(strings.Repeat("z", 40), DefaultLifetime, func() time.Time { return issuedAt })
	if _, err := foreignManager.Verify(issuedToken.Value); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("token from another secret: got %v", err)
	}

	unsignedToken, _ := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{Subject: uuid.NewString()}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if _, err := newTestManager(t, issuedAt).Verify(unsignedToken); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("alg=none token: got %v", err)
	}
}

func TestNewManagerRejectsShortSecrets(t *testing.T) {
	if _, err := NewManager("short", DefaultLifetime, time.Now); !errors.Is(err, ErrSecretTooShort) {
		t.Fatalf("got %v, want ErrSecretTooShort", err)
	}
}
