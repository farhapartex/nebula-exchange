package outbox

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/email"
)

const DefaultMaximumAttempts = 10

type Enqueuer interface {
	Enqueue(ctx context.Context, templateName string, message email.Message) error
}

type RepositoryEnqueuer struct {
	repository Repository
	now        func() time.Time
}

func NewEnqueuer(repository Repository, now func() time.Time) *RepositoryEnqueuer {
	return &RepositoryEnqueuer{repository: repository, now: now}
}

func (enqueuer *RepositoryEnqueuer) Enqueue(ctx context.Context, templateName string, message email.Message) error {
	emailID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	return enqueuer.repository.Insert(ctx, &OutboxEmail{
		ID:             emailID,
		Template:       templateName,
		RecipientEmail: message.To.Email,
		RecipientName:  message.To.Name,
		Subject:        message.Subject,
		HTMLBody:       message.HTMLBody,
		TextBody:       message.TextBody,
		Status:         EmailStatusPending,
		MaxAttempts:    DefaultMaximumAttempts,
		NextAttemptAt:  enqueuer.now().UTC(),
	})
}
