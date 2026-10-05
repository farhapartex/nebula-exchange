package models

import "github.com/farhapartex/nebula-exchange/backend/internal/platform/database"

type LevelEnemy struct {
	LevelID   string                `gorm:"primaryKey"`
	Wave      int                   `gorm:"primaryKey"`
	EnemyID   string                `gorm:"not null"`
	Modifiers database.JSONDocument `gorm:"type:jsonb;not null"`
	IntroLine *string

	Enemy Enemy `gorm:"foreignKey:EnemyID"`
}

func (LevelEnemy) TableName() string {
	return "level_enemies"
}
