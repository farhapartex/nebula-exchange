package maintenance_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"nebula-exchange/backend/internal/maintenance"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/logger"
)

func TestCleanupRemovesOnlyStaleRows(t *testing.T) {
	pool := databasetest.NewPool(t)
	ctx := context.Background()
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

	insertUser := func(isActive bool, createdAt time.Time) uuid.UUID {
		userID := uuid.New()
		var activatedAt *time.Time
		if isActive {
			activatedAt = &createdAt
		}
		_, err := pool.Exec(ctx, `INSERT INTO users (id, email, username, password_hash, terms_accepted_at, is_active, activated_at, created_at)
			VALUES ($1, $2, $3, 'hash', $4, $5, $6, $4)`, userID, userID.String()+"@nebula.test", "u_"+userID.String()[:8], createdAt, isActive, activatedAt)
		if err != nil {
			t.Fatalf("insert user: %v", err)
		}
		return userID
	}
	staleInactiveUser := insertUser(false, now.Add(-8*24*time.Hour))
	recentInactiveUser := insertUser(false, now.Add(-2*24*time.Hour))
	oldActiveUser := insertUser(true, now.Add(-90*24*time.Hour))

	_, _ = pool.Exec(ctx, "INSERT INTO idempotency_keys (scope, idempotency_key, request_hash, status, created_at) VALUES ('anonymous', 'old-key-0001', '\\x00', 'completed', $1), ('anonymous', 'new-key-0001', '\\x00', 'completed', $2)", now.Add(-48*time.Hour), now.Add(-time.Hour))
	_, _ = pool.Exec(ctx, `INSERT INTO email_outbox (id, template, recipient_email, subject, html_body, text_body, status, sent_at)
		VALUES ($1, 't', 'a@nebula.test', 's', 'h', 't', 'sent', $2), ($3, 't', 'b@nebula.test', 's', 'h', 't', 'failed', NULL)`,
		uuid.New(), now.Add(-40*24*time.Hour), uuid.New())

	cleanupJob := maintenance.NewCleanupJob(pool, logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true), func() time.Time { return now })
	if err := cleanupJob.Run(ctx); err != nil {
		t.Fatalf("run cleanup: %v", err)
	}

	userExists := func(userID uuid.UUID) bool {
		var matchCount int
		_ = pool.QueryRow(ctx, "SELECT count(*) FROM users WHERE id = $1", userID).Scan(&matchCount)
		return matchCount == 1
	}
	if userExists(staleInactiveUser) {
		t.Fatal("accounts never activated within 7 days must be deleted")
	}
	if !userExists(recentInactiveUser) || !userExists(oldActiveUser) {
		t.Fatal("recent unactivated and all activated accounts must be kept")
	}

	var remainingKeys, remainingEmails int
	_ = pool.QueryRow(ctx, "SELECT count(*) FROM idempotency_keys").Scan(&remainingKeys)
	_ = pool.QueryRow(ctx, "SELECT count(*) FROM email_outbox").Scan(&remainingEmails)
	if remainingKeys != 1 {
		t.Fatalf("only idempotency keys younger than 24 hours stay, got %d", remainingKeys)
	}
	if remainingEmails != 1 {
		t.Fatalf("old sent emails go, failed ones stay, got %d", remainingEmails)
	}
}
