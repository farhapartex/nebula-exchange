package models

import (
	"time"

	"github.com/google/uuid"
)

type CoinClawbackFlag struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID `gorm:"type:uuid;not null"`
	PaymentID       uuid.UUID `gorm:"type:uuid;not null"`
	PackCoins       int64     `gorm:"not null"`
	ClawedBackCoins int64     `gorm:"not null"`
	MissingCoins    int64     `gorm:"not null"`
	ReviewedAt      *time.Time
	CreatedAt       time.Time `gorm:"not null"`
}

func (CoinClawbackFlag) TableName() string {
	return "coin_clawback_flags"
}
