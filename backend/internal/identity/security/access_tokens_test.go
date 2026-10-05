package security

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const testSecret = "test-secret-that-is-long-enough-for-hs256"

func newTestAccessTokens(t *testing.T, now time.Time) *AccessTokens {
	t.Helper()
	manager, err := NewAccessTokens(testSecret, DefaultAccessTokenLifetime, func() time.Time { return now })
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	return manager
}

func TestIssueAndVerifyRoundTrip(t *testing.T) {
	issuedAt := time.Date(2026, 9, 30, 10, 0, 0, 0, time.UTC)
	accessTokens := newTestAccessTokens(t, issuedAt)
	userID := uuid.New()

	issuedToken, err := accessTokens.Issue(userID)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	if !issuedToken.ExpiresAt.Equal(issuedAt.Add(15 * time.Minute)) {
		t.Fatalf("got expiry %v", issuedToken.ExpiresAt)
	}

	verifiedUserID, err := accessTokens.Verify(issuedToken.Value)
	if err != nil || verifiedUserID != userID {
		t.Fatalf("got %v, %v, want %v", verifiedUserID, err, userID)
	}
}

func TestVerifyRejectsExpiredTamperedAndForeignTokens(t *testing.T) {
	issuedAt := time.Date(2026, 9, 30, 10, 0, 0, 0, time.UTC)
	issuedToken, _ := newTestAccessTokens(t, issuedAt).Issue(uuid.New())

	laterAccessTokens := newTestAccessTokens(t, issuedAt.Add(16*time.Minute))
	if _, err := laterAccessTokens.Verify(issuedToken.Value); !errors.Is(err, ErrAccessTokenInvalid) {
		t.Fatalf("expired token: got %v", err)
	}

	tamperedToken := issuedToken.Value[:len(issuedToken.Value)-2] + "xx"
	if _, err := newTestAccessTokens(t, issuedAt).Verify(tamperedToken); !errors.Is(err, ErrAccessTokenInvalid) {
		t.Fatalf("tampered token: got %v", err)
	}

	foreignAccessTokens, _ := NewAccessTokens(strings.Repeat("z", 40), DefaultAccessTokenLifetime, func() time.Time { return issuedAt })
	if _, err := foreignAccessTokens.Verify(issuedToken.Value); !errors.Is(err, ErrAccessTokenInvalid) {
		t.Fatalf("token from another secret: got %v", err)
	}

	unsignedToken, _ := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{Subject: uuid.NewString()}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if _, err := newTestAccessTokens(t, issuedAt).Verify(unsignedToken); !errors.Is(err, ErrAccessTokenInvalid) {
		t.Fatalf("alg=none token: got %v", err)
	}
}

func TestNewAccessTokensRejectsShortSecrets(t *testing.T) {
	if _, err := NewAccessTokens("short", DefaultAccessTokenLifetime, time.Now); !errors.Is(err, ErrAccessTokenSecretTooShort) {
		t.Fatalf("got %v, want ErrAccessTokenSecretTooShort", err)
	}
}
