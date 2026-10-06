package models

import (
	"time"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type StripeWebhookEvent struct {
	ID          string                `gorm:"primaryKey"`
	Type        string                `gorm:"not null"`
	Payload     database.JSONDocument `gorm:"type:jsonb;not null"`
	ReceivedAt  time.Time             `gorm:"not null"`
	ProcessedAt *time.Time
}

func (StripeWebhookEvent) TableName() string {
	return "stripe_webhook_events"
}
