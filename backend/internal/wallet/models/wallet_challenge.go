package models

import (
	"time"

	"github.com/google/uuid"
)

type WalletChallenge struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	Address   string    `gorm:"not null"`
	ChainID   int64     `gorm:"not null"`
	Nonce     string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time
}

func (WalletChallenge) TableName() string {
	return "wallet_challenges"
}
