package identity

import (
	"time"

	"github.com/google/uuid"
)

type ActivationToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	TokenHash []byte    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time
}

func (ActivationToken) TableName() string {
	return "activation_tokens"
}
