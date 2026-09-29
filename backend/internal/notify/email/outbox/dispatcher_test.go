package outbox_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/notify/email"
	"nebula-exchange/backend/internal/notify/email/emailtest"
	"nebula-exchange/backend/internal/notify/email/outbox"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/logger"
)

type testClock struct {
	currentTime atomic.Pointer[time.Time]
}

func newTestClock(start time.Time) *testClock {
	clock := &testClock{}
	clock.set(start)
	return clock
}

func (clock *testClock) set(newTime time.Time) { clock.currentTime.Store(&newTime) }
func (clock *testClock) now() time.Time        { return *clock.currentTime.Load() }
func (clock *testClock) advance(duration time.Duration) {
	clock.set(clock.now().Add(duration))
}

type outboxTestHarness struct {
	pool   *pgxpool.Pool
	sender *emailtest.RecordingSender
	clock  *testClock
}

func newOutboxTestHarness(t *testing.T) *outboxTestHarness {
	t.Helper()
	pool := databasetest.NewPool(t)
	clock := newTestClock(time.Now().UTC().Truncate(time.Second))
	return &outboxTestHarness{pool: pool, sender: emailtest.NewRecordingSender(), clock: clock}
}

func (harness *outboxTestHarness) newDispatcher() *outbox.Dispatcher {
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	return outbox.NewDispatcher(harness.pool, harness.sender, testLogger, outbox.DispatcherOptions{
		BatchSize: 20,
		Now:       harness.clock.now,
	})
}

func (harness *outboxTestHarness) enqueue(t *testing.T, recipientEmail string) {
	t.Helper()
	_, err := outbox.NewQueue().Enqueue(context.Background(), harness.pool, email.TemplateAccountActivation, email.Message{
		To:       email.Address{Email: recipientEmail},
		Subject:  "Activate",
		HTMLBody: "<p>activate</p>",
		TextBody: "activate",
	})
	if err != nil {
		t.Fatalf("enqueue email: %v", err)
	}
	_, err = harness.pool.Exec(context.Background(), "UPDATE email_outbox SET next_attempt_at = $1", harness.clock.now())
	if err != nil {
		t.Fatalf("align next attempt with test clock: %v", err)
	}
}

type outboxRow struct {
	status        string
	attempts      int
	nextAttemptAt time.Time
	lastError     *string
}

func (harness *outboxTestHarness) onlyRow(t *testing.T) outboxRow {
	t.Helper()
	var row outboxRow
	err := harness.pool.QueryRow(context.Background(),
		"SELECT status, attempts, next_attempt_at, last_error FROM email_outbox",
	).Scan(&row.status, &row.attempts, &row.nextAttemptAt, &row.lastError)
	if err != nil {
		t.Fatalf("load outbox row: %v", err)
	}
	return row
}

func TestDispatcherSendsPendingEmailAndMarksItSent(t *testing.T) {
	harness := newOutboxTestHarness(t)
	harness.enqueue(t, "pilot@nebula.test")

	dispatchedCount, err := harness.newDispatcher().DispatchDueBatch(context.Background())

	if err != nil || dispatchedCount != 1 {
		t.Fatalf("got %d, %v, want one dispatched email", dispatchedCount, err)
	}
	if sentMessages := harness.sender.SentMessages(); len(sentMessages) != 1 || sentMessages[0].To.Email != "pilot@nebula.test" {
		t.Fatalf("unexpected sent messages %+v", sentMessages)
	}
	if row := harness.onlyRow(t); row.status != "sent" || row.attempts != 1 {
		t.Fatalf("unexpected row %+v", row)
	}
	if secondCount, _ := harness.newDispatcher().DispatchDueBatch(context.Background()); secondCount != 0 {
		t.Fatal("a sent email must not be dispatched again")
	}
}

func TestDispatcherRetriesWithBackoffAfterFailure(t *testing.T) {
	harness := newOutboxTestHarness(t)
	harness.enqueue(t, "pilot@nebula.test")
	harness.sender.FailWith(errors.New("smtp: 421 service not available"))
	dispatcher := harness.newDispatcher()
	failedAt := harness.clock.now()

	dispatcher.DispatchDueBatch(context.Background())

	failedRow := harness.onlyRow(t)
	if failedRow.status != "pending" || failedRow.attempts != 1 || failedRow.lastError == nil {
		t.Fatalf("unexpected row after failure %+v", failedRow)
	}
	if !failedRow.nextAttemptAt.Equal(failedAt.Add(30 * time.Second)) {
		t.Fatalf("got next attempt %v, want 30 seconds after failure", failedRow.nextAttemptAt)
	}

	harness.clock.advance(29 * time.Second)
	if dispatchedCount, _ := dispatcher.DispatchDueBatch(context.Background()); dispatchedCount != 0 {
		t.Fatal("email must wait for its backoff before retrying")
	}

	harness.sender.FailWith(nil)
	harness.clock.advance(2 * time.Second)
	if dispatchedCount, _ := dispatcher.DispatchDueBatch(context.Background()); dispatchedCount != 1 {
		t.Fatal("email must retry once the backoff has passed")
	}
	if recoveredRow := harness.onlyRow(t); recoveredRow.status != "sent" || recoveredRow.attempts != 2 || recoveredRow.lastError != nil {
		t.Fatalf("unexpected row after recovery %+v", recoveredRow)
	}
}

func TestDispatcherGivesUpAfterTenAttempts(t *testing.T) {
	harness := newOutboxTestHarness(t)
	harness.enqueue(t, "pilot@nebula.test")
	harness.sender.FailWith(errors.New("smtp: mailbox unavailable"))
	dispatcher := harness.newDispatcher()

	for attemptNumber := 1; attemptNumber <= outbox.DefaultMaximumAttempts; attemptNumber++ {
		if dispatchedCount, _ := dispatcher.DispatchDueBatch(context.Background()); dispatchedCount != 1 {
			t.Fatalf("attempt %d was not dispatched", attemptNumber)
		}
		harness.clock.advance(2 * time.Hour)
	}

	if row := harness.onlyRow(t); row.status != "failed" || row.attempts != outbox.DefaultMaximumAttempts {
		t.Fatalf("got %+v, want failed after %d attempts", row, outbox.DefaultMaximumAttempts)
	}
	if dispatchedCount, _ := dispatcher.DispatchDueBatch(context.Background()); dispatchedCount != 0 {
		t.Fatal("a failed email must not be retried again")
	}
}

func TestConcurrentDispatchersSendEachEmailExactlyOnce(t *testing.T) {
	harness := newOutboxTestHarness(t)
	const queuedEmailCount = 60
	for emailIndex := 0; emailIndex < queuedEmailCount; emailIndex++ {
		harness.enqueue(t, fmt.Sprintf("pilot%02d@nebula.test", emailIndex))
	}

	var waitGroup sync.WaitGroup
	for dispatcherIndex := 0; dispatcherIndex < 4; dispatcherIndex++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			dispatcher := harness.newDispatcher()
			for {
				dispatchedCount, err := dispatcher.DispatchDueBatch(context.Background())
				if err != nil {
					t.Errorf("dispatch: %v", err)
					return
				}
				if dispatchedCount == 0 {
					return
				}
			}
		}()
	}
	waitGroup.Wait()

	sendCountByRecipient := map[string]int{}
	for _, sentMessage := range harness.sender.SentMessages() {
		sendCountByRecipient[sentMessage.To.Email]++
	}
	if len(sendCountByRecipient) != queuedEmailCount {
		t.Fatalf("sent to %d recipients, want %d", len(sendCountByRecipient), queuedEmailCount)
	}
	for recipientEmail, sendCount := range sendCountByRecipient {
		if sendCount != 1 {
			t.Fatalf("%s received %d emails, want exactly 1", recipientEmail, sendCount)
		}
	}
}

func TestExpiredClaimIsPickedUpAgain(t *testing.T) {
	harness := newOutboxTestHarness(t)
	harness.enqueue(t, "pilot@nebula.test")
	_, err := harness.pool.Exec(context.Background(),
		"UPDATE email_outbox SET claimed_until = $1", harness.clock.now().Add(time.Minute))
	if err != nil {
		t.Fatalf("simulate a crashed dispatcher's claim: %v", err)
	}
	dispatcher := harness.newDispatcher()

	if dispatchedCount, _ := dispatcher.DispatchDueBatch(context.Background()); dispatchedCount != 0 {
		t.Fatal("an email claimed by another dispatcher must be skipped")
	}
	harness.clock.advance(61 * time.Second)
	if dispatchedCount, _ := dispatcher.DispatchDueBatch(context.Background()); dispatchedCount != 1 {
		t.Fatal("an email with an expired claim must be dispatched")
	}
}
