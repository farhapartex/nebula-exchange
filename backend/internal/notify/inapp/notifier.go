package inapp

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"nebula-exchange/backend/internal/notify/email"
	"nebula-exchange/backend/internal/notify/email/outbox"
	"nebula-exchange/backend/internal/notify/inapp/inappstore"
	"nebula-exchange/backend/internal/users"
	"nebula-exchange/backend/internal/users/usersstore"
)

type Kind string

const (
	KindPaymentSucceeded Kind = "payment_succeeded"
	KindPaymentNeedsCare Kind = "payment_needs_attention"
	KindMissionCompleted Kind = "mission_completed"
	KindCraftCompleted   Kind = "craft_completed"
)

type Notice struct {
	UserID        uuid.UUID
	Kind          Kind
	Title         string
	Body          string
	Link          string
	AlsoSendEmail bool
}

type Database interface {
	inappstore.DBTX
	usersstore.DBTX
}

type Dependencies struct {
	Users           *users.Repository
	EmailTemplates  *email.TemplateRenderer
	EmailQueue      *outbox.Queue
	FrontendBaseURL string
}

type Notifier struct {
	dependencies Dependencies
}

func NewNotifier(dependencies Dependencies) *Notifier {
	return &Notifier{dependencies: dependencies}
}

func (notifier *Notifier) Notify(ctx context.Context, database Database, notice Notice) error {
	notificationID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	if _, err := inappstore.New(database).InsertNotification(ctx, inappstore.InsertNotificationParams{
		ID:     notificationID,
		UserID: notice.UserID,
		Kind:   string(notice.Kind),
		Title:  notice.Title,
		Body:   notice.Body,
		Link:   notice.Link,
	}); err != nil {
		return fmt.Errorf("insert notification: %w", err)
	}
	if !notice.AlsoSendEmail || notifier.dependencies.EmailQueue == nil {
		return nil
	}
	return notifier.sendEmail(ctx, database, notice)
}

func (notifier *Notifier) sendEmail(ctx context.Context, database Database, notice Notice) error {
	recipient, isFound, err := notifier.dependencies.Users.FindByID(ctx, database, notice.UserID)
	if err != nil || !isFound {
		return err
	}
	htmlBody, textBody, err := notifier.dependencies.EmailTemplates.Render(email.TemplateGameNotice, struct {
		Username, Title, Body, Link string
	}{recipient.Username, notice.Title, notice.Body, notifier.dependencies.FrontendBaseURL + notice.Link})
	if err != nil {
		return err
	}
	_, err = notifier.dependencies.EmailQueue.Enqueue(ctx, database, email.TemplateGameNotice, email.Message{
		To:       email.Address{Name: recipient.Username, Email: recipient.Email},
		Subject:  notice.Title,
		HTMLBody: htmlBody,
		TextBody: textBody,
	})
	return err
}

type Notification struct {
	ID        uuid.UUID  `json:"id"`
	Kind      string     `json:"kind"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Link      string     `json:"link"`
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at"`
}

func notificationFromRow(notificationRow inappstore.Notification) Notification {
	return Notification{
		ID:        notificationRow.ID,
		Kind:      notificationRow.Kind,
		Title:     notificationRow.Title,
		Body:      notificationRow.Body,
		Link:      notificationRow.Link,
		CreatedAt: notificationRow.CreatedAt,
		ReadAt:    notificationRow.ReadAt,
	}
}
