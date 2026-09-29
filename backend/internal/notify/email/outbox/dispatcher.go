package outbox

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/notify/email"
	"nebula-exchange/backend/internal/notify/email/outbox/outboxstore"
)

const maximumStoredErrorLength = 1000

type DispatcherOptions struct {
	PollInterval  time.Duration
	BatchSize     int32
	ClaimDuration time.Duration
	Now           func() time.Time
}

type Dispatcher struct {
	pool    *pgxpool.Pool
	sender  email.Sender
	logger  *slog.Logger
	options DispatcherOptions
}

func NewDispatcher(pool *pgxpool.Pool, sender email.Sender, logger *slog.Logger, options DispatcherOptions) *Dispatcher {
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
	return &Dispatcher{pool: pool, sender: sender, logger: logger, options: options}
}

func (dispatcher *Dispatcher) Run(ctx context.Context) {
	dispatcher.logger.Info("email dispatcher started", slog.Duration("poll_interval", dispatcher.options.PollInterval))
	pollTicker := time.NewTicker(dispatcher.options.PollInterval)
	defer pollTicker.Stop()

	for {
		if _, err := dispatcher.DispatchDueBatch(ctx); err != nil && !errors.Is(err, context.Canceled) {
			dispatcher.logger.Error("dispatch email batch", slog.String("error", err.Error()))
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
	claimedEmails, err := outboxstore.New(dispatcher.pool).ClaimDueEmails(ctx, outboxstore.ClaimDueEmailsParams{
		ClaimedUntil: now.Add(dispatcher.options.ClaimDuration),
		Now:          now,
		BatchSize:    dispatcher.options.BatchSize,
	})
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

func (dispatcher *Dispatcher) deliver(ctx context.Context, claimedEmail outboxstore.ClaimDueEmailsRow) {
	sendError := dispatcher.sender.Send(ctx, email.Message{
		To:       email.Address{Name: claimedEmail.RecipientName, Email: claimedEmail.RecipientEmail},
		Subject:  claimedEmail.Subject,
		HTMLBody: claimedEmail.HtmlBody,
		TextBody: claimedEmail.TextBody,
	})

	persistContext := context.WithoutCancel(ctx)
	queries := outboxstore.New(dispatcher.pool)
	now := dispatcher.options.Now().UTC()

	if sendError == nil {
		if err := queries.MarkEmailSent(persistContext, outboxstore.MarkEmailSentParams{ID: claimedEmail.ID, Now: now}); err != nil {
			dispatcher.logger.Error("mark email sent", slog.String("email_id", claimedEmail.ID.String()), slog.String("error", err.Error()))
		}
		return
	}

	failedAttemptCount := int(claimedEmail.Attempts) + 1
	isFinalAttempt := failedAttemptCount >= int(claimedEmail.MaxAttempts)
	err := queries.RecordEmailFailure(persistContext, outboxstore.RecordEmailFailureParams{
		ID:            claimedEmail.ID,
		NextAttemptAt: now.Add(RetryDelay(failedAttemptCount)),
		LastError:     truncate(sendError.Error(), maximumStoredErrorLength),
		Now:           now,
	})
	if err != nil {
		dispatcher.logger.Error("record email failure", slog.String("email_id", claimedEmail.ID.String()), slog.String("error", err.Error()))
	}

	logLevel := slog.LevelWarn
	if isFinalAttempt {
		logLevel = slog.LevelError
	}
	dispatcher.logger.Log(ctx, logLevel, "email delivery failed",
		slog.String("email_id", claimedEmail.ID.String()),
		slog.Int("attempt", failedAttemptCount),
		slog.Bool("gave_up", isFinalAttempt),
		slog.String("error", sendError.Error()),
	)
}

func truncate(text string, maximumLength int) string {
	if len(text) <= maximumLength {
		return text
	}
	return text[:maximumLength]
}
