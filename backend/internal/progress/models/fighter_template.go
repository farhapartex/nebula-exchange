package models

import (
	"time"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type FighterTemplate struct {
	ID            string                `gorm:"primaryKey"`
	Name          string                `gorm:"not null"`
	Title         string                `gorm:"not null"`
	StartingLevel int                   `gorm:"not null"`
	Stats         database.JSONDocument `gorm:"type:jsonb;not null"`
	Look          database.JSONDocument `gorm:"type:jsonb;not null"`
	IsDefault     bool                  `gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (FighterTemplate) TableName() string {
	return "fighter_templates"
}
