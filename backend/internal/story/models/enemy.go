package models

import (
	"time"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type Enemy struct {
	ID        string                `gorm:"primaryKey"`
	Name      string                `gorm:"not null"`
	Title     string                `gorm:"not null"`
	Stats     database.JSONDocument `gorm:"type:jsonb;not null"`
	Brain     database.JSONDocument `gorm:"type:jsonb;not null"`
	Look      database.JSONDocument `gorm:"type:jsonb;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Enemy) TableName() string {
	return "enemies"
}
