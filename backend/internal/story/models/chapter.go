package models

import "time"

type Chapter struct {
	ID          string `gorm:"primaryKey"`
	Number      int    `gorm:"not null"`
	Title       string `gorm:"not null"`
	Summary     string `gorm:"not null"`
	IsFree      bool   `gorm:"not null"`
	PriceCoins  *int64
	IsPublished bool `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Chapter) TableName() string {
	return "chapters"
}
