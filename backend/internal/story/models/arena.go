package models

import (
	"time"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type Arena struct {
	ID        string                `gorm:"primaryKey"`
	Name      string                `gorm:"not null"`
	Width     int                   `gorm:"not null"`
	FloorY    int                   `gorm:"column:floor_y;not null"`
	Stage     database.JSONDocument `gorm:"type:jsonb;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Arena) TableName() string {
	return "arenas"
}
