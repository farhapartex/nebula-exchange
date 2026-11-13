package models

import (
	"time"

	"github.com/google/uuid"
)

type CoinBalance struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Coins     int64     `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (CoinBalance) TableName() string {
	return "coin_balances"
}
