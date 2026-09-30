package payments_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"nebula-exchange/backend/internal/payments"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/users"
)

func TestExpiryJobExpiresOnlyOverduePendingPayments(t *testing.T) {
	harness := newPaymentsHarness(t)
	player := harness.createPlayer(t, users.StatusActive)
	overdue := harness.createPayment(t, player, `{"purpose":"TOPUP","method":"card","amount_nc":"10"}`)
	fresh := harness.createPayment(t, player, `{"purpose":"TOPUP","method":"card","amount_nc":"5"}`)
	settled := harness.createPayment(t, player, `{"purpose":"TOPUP","method":"card","amount_nc":"25"}`)
	harness.settle(t, settled, "evt_settled")
	harness.pool.Exec(context.Background(), "UPDATE payments SET expires_at = now() - interval '1 minute' WHERE id = ANY($1)",
		[]any{overdue.ID, settled.ID})

	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	if err := payments.NewExpiryJob(harness.pool, testLogger, time.Now).Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}

	expectedStatuses := map[string]string{overdue.ID.String(): "EXPIRED", fresh.ID.String(): "PENDING", settled.ID.String(): "SUCCEEDED"}
	for paymentID, expectedStatus := range expectedStatuses {
		var status string
		harness.pool.QueryRow(context.Background(), "SELECT status FROM payments WHERE id = $1", paymentID).Scan(&status)
		if status != expectedStatus {
			t.Fatalf("payment %s is %s, want %s", paymentID, status, expectedStatus)
		}
	}
	if outcome := harness.settle(t, overdue, "evt_late"); outcome.WasSettled {
		t.Fatal("an expired payment must not be settled by a late event")
	}
}
