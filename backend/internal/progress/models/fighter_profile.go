package models

import (
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type FighterProfile struct {
	UserID       uuid.UUID             `gorm:"type:uuid;primaryKey"`
	TemplateID   string                `gorm:"not null"`
	FighterLevel int                   `gorm:"not null"`
	Experience   int                   `gorm:"not null"`
	Wins         int                   `gorm:"not null"`
	Losses       int                   `gorm:"not null"`
	Stats        database.JSONDocument `gorm:"type:jsonb;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Template FighterTemplate `gorm:"foreignKey:TemplateID"`
}

func (FighterProfile) TableName() string {
	return "fighter_profiles"
}
