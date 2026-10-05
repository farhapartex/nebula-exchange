package outbox

import (
	"time"

	"github.com/google/uuid"
)

type EmailStatus string

const (
	EmailStatusPending EmailStatus = "PENDING"
	EmailStatusSent    EmailStatus = "SENT"
	EmailStatusFailed  EmailStatus = "FAILED"
)

type OutboxEmail struct {
	ID             uuid.UUID   `gorm:"type:uuid;primaryKey"`
	Template       string      `gorm:"not null"`
	RecipientEmail string      `gorm:"type:citext;not null"`
	RecipientName  string      `gorm:"not null"`
	Subject        string      `gorm:"not null"`
	HTMLBody       string      `gorm:"column:html_body;not null"`
	TextBody       string      `gorm:"not null"`
	Status         EmailStatus `gorm:"type:email_status;not null"`
	Attempts       int         `gorm:"not null"`
	MaxAttempts    int         `gorm:"not null"`
	NextAttemptAt  time.Time   `gorm:"not null"`
	ClaimedUntil   *time.Time
	LastError      *string
	SentAt         *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (OutboxEmail) TableName() string {
	return "email_outbox"
}
