package outbox_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database/databasetest"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/email"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/email/outbox"
)

type recordingSender struct {
	mutex        sync.Mutex
	sentMessages []email.Message
	failure      error
}

func (sender *recordingSender) Send(_ context.Context, message email.Message) error {
	sender.mutex.Lock()
	defer sender.mutex.Unlock()
	if sender.failure != nil {
		return sender.failure
	}
	sender.sentMessages = append(sender.sentMessages, message)
	return nil
}

func newDispatchHarness(t *testing.T, sender *recordingSender, now func() time.Time) (*outbox.RepositoryEnqueuer, *outbox.Dispatcher, *outbox.GormRepository) {
	t.Helper()
	repository := outbox.NewGormRepository(databasetest.Open(t))
	quietLogger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	return outbox.NewEnqueuer(repository, now), outbox.NewDispatcher(repository, sender, quietLogger, outbox.DispatcherOptions{Now: now}), repository
}

func activationMessage() email.Message {
	return email.Message{To: email.Address{Name: "the_boy", Email: "boy@streetborn.test"}, Subject: "Activate", HTMLBody: "<p>hi</p>", TextBody: "hi"}
}

func TestQueuedEmailIsSentOnce(t *testing.T) {
	sender := &recordingSender{}
	enqueuer, dispatcher, _ := newDispatchHarness(t, sender, time.Now)

	if err := enqueuer.Enqueue(context.Background(), "account_activation", activationMessage()); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	firstBatch, err := dispatcher.DispatchDueBatch(context.Background())
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	secondBatch, _ := dispatcher.DispatchDueBatch(context.Background())

	if firstBatch != 1 || secondBatch != 0 || len(sender.sentMessages) != 1 {
		t.Fatalf("got batches %d and %d with %d sent, want one email sent once", firstBatch, secondBatch, len(sender.sentMessages))
	}
	if sender.sentMessages[0].To.Email != "boy@streetborn.test" {
		t.Fatalf("sent to %q", sender.sentMessages[0].To.Email)
	}
}

func TestFailedEmailWaitsForItsRetryTime(t *testing.T) {
	sender := &recordingSender{failure: errors.New("smtp is down")}
	currentTime := time.Now()
	enqueuer, dispatcher, _ := newDispatchHarness(t, sender, func() time.Time { return currentTime })

	if err := enqueuer.Enqueue(context.Background(), "account_activation", activationMessage()); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if _, err := dispatcher.DispatchDueBatch(context.Background()); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	sender.failure = nil
	immediateRetry, _ := dispatcher.DispatchDueBatch(context.Background())
	currentTime = currentTime.Add(outbox.RetryDelay(1) + time.Second)
	laterRetry, _ := dispatcher.DispatchDueBatch(context.Background())

	if immediateRetry != 0 || laterRetry != 1 || len(sender.sentMessages) != 1 {
		t.Fatalf("got %d then %d with %d sent, want the retry only after the delay", immediateRetry, laterRetry, len(sender.sentMessages))
	}
}
