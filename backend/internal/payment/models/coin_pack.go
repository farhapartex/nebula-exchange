package models

import "time"

type CoinPack struct {
	ID         string `gorm:"primaryKey"`
	Name       string `gorm:"not null"`
	Coins      int64  `gorm:"not null"`
	PriceCents int64  `gorm:"not null"`
	SortOrder  int    `gorm:"not null"`
	IsActive   bool   `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (CoinPack) TableName() string {
	return "coin_packs"
}
