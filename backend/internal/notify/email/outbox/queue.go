package outbox

import (
	"context"

	"github.com/google/uuid"

	"nebula-exchange/backend/internal/notify/email"
	"nebula-exchange/backend/internal/notify/email/outbox/outboxstore"
)

const DefaultMaximumAttempts = 10

type Queue struct {
	maximumAttempts int32
}

func NewQueue() *Queue {
	return &Queue{maximumAttempts: DefaultMaximumAttempts}
}

func (queue *Queue) Enqueue(ctx context.Context, database outboxstore.DBTX, templateName email.TemplateName, message email.Message) (uuid.UUID, error) {
	outboxEmailID, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}
	err = outboxstore.New(database).EnqueueEmail(ctx, outboxstore.EnqueueEmailParams{
		ID:             outboxEmailID,
		Template:       string(templateName),
		RecipientEmail: message.To.Email,
		RecipientName:  message.To.Name,
		Subject:        message.Subject,
		HtmlBody:       message.HTMLBody,
		TextBody:       message.TextBody,
		MaxAttempts:    queue.maximumAttempts,
	})
	return outboxEmailID, err
}
