package activation

import (
	"context"

	"nebula-exchange/backend/internal/notify/email"
	"nebula-exchange/backend/internal/notify/email/outbox"
	"nebula-exchange/backend/internal/notify/email/outbox/outboxstore"
)

type Mailer struct {
	composer *EmailComposer
	queue    *outbox.Queue
}

func NewMailer(composer *EmailComposer, queue *outbox.Queue) *Mailer {
	return &Mailer{composer: composer, queue: queue}
}

func (mailer *Mailer) QueueActivationEmail(ctx context.Context, database outboxstore.DBTX, recipientEmail, username string, issuedToken IssuedToken) error {
	activationMessage, err := mailer.composer.Compose(email.Address{Name: username, Email: recipientEmail}, username, issuedToken)
	if err != nil {
		return err
	}
	_, err = mailer.queue.Enqueue(ctx, database, email.TemplateAccountActivation, activationMessage)
	return err
}
