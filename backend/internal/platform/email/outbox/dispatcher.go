package outbox

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/email"
)

const maximumStoredErrorLength = 1000

type DispatcherOptions struct {
	PollInterval  time.Duration
	BatchSize     int
	ClaimDuration time.Duration
	Now           func() time.Time
}

type Dispatcher struct {
	repository Repository
	sender     email.Sender
	logger     *slog.Logger
	options    DispatcherOptions
}

func NewDispatcher(repository Repository, sender email.Sender, logger *slog.Logger, options DispatcherOptions) *Dispatcher {
	if options.PollInterval == 0 {
		options.PollInterval = 2 * time.Second
	}
	if options.BatchSize == 0 {
		options.BatchSize = 20
	}
	if options.ClaimDuration == 0 {
		options.ClaimDuration = time.Minute
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Dispatcher{repository: repository, sender: sender, logger: logger, options: options}
}

func (dispatcher *Dispatcher) Run(ctx context.Context) {
	dispatcher.logger.Info("email dispatcher started", slog.Duration("poll_interval", dispatcher.options.PollInterval))
	pollTicker := time.NewTicker(dispatcher.options.PollInterval)
	defer pollTicker.Stop()
	for {
		if _, err := dispatcher.DispatchDueBatch(ctx); err != nil && !errors.Is(err, context.Canceled) {
			dispatcher.logger.Error("dispatch email batch", slog.Any("error", err))
		}
		select {
		case <-ctx.Done():
			dispatcher.logger.Info("email dispatcher stopped")
			return
		case <-pollTicker.C:
		}
	}
}

func (dispatcher *Dispatcher) DispatchDueBatch(ctx context.Context) (int, error) {
	now := dispatcher.options.Now().UTC()
	claimedEmails, err := dispatcher.repository.ClaimDue(ctx, now, now.Add(dispatcher.options.ClaimDuration), dispatcher.options.BatchSize)
	if err != nil {
		return 0, err
	}
	for _, claimedEmail := range claimedEmails {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		dispatcher.deliver(ctx, claimedEmail)
	}
	return len(claimedEmails), nil
}

func (dispatcher *Dispatcher) deliver(ctx context.Context, claimedEmail OutboxEmail) {
	sendError := dispatcher.sender.Send(ctx, email.Message{
		To:       email.Address{Name: claimedEmail.RecipientName, Email: claimedEmail.RecipientEmail},
		Subject:  claimedEmail.Subject,
		HTMLBody: claimedEmail.HTMLBody,
		TextBody: claimedEmail.TextBody,
	})
	persistContext := context.WithoutCancel(ctx)
	now := dispatcher.options.Now().UTC()
	if sendError == nil {
		if err := dispatcher.repository.MarkSent(persistContext, claimedEmail.ID, now); err != nil {
			dispatcher.logger.Error("mark email sent", slog.String("email_id", claimedEmail.ID.String()), slog.Any("error", err))
		}
		return
	}

	failedAttemptCount := claimedEmail.Attempts + 1
	isFinalAttempt := failedAttemptCount >= claimedEmail.MaxAttempts
	err := dispatcher.repository.RecordFailure(persistContext, claimedEmail, now.Add(RetryDelay(failedAttemptCount)), truncate(sendError.Error(), maximumStoredErrorLength), now)
	if err != nil {
		dispatcher.logger.Error("record email failure", slog.String("email_id", claimedEmail.ID.String()), slog.Any("error", err))
	}
	logLevel := slog.LevelWarn
	if isFinalAttempt {
		logLevel = slog.LevelError
	}
	dispatcher.logger.Log(ctx, logLevel, "email delivery failed",
		slog.String("email_id", claimedEmail.ID.String()),
		slog.Int("attempt", failedAttemptCount),
		slog.Bool("gave_up", isFinalAttempt),
		slog.Any("error", sendError),
	)
}

func truncate(text string, maximumLength int) string {
	if len(text) <= maximumLength {
		return text
	}
	return text[:maximumLength]
}
