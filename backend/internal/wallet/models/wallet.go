package models

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	Address   string    `gorm:"not null"`
	ChainID   int64     `gorm:"not null"`
	LinkedAt  time.Time `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Wallet) TableName() string {
	return "wallets"
}
