package models

import (
	"time"

	"github.com/google/uuid"
)

type ToolMasteryEvent struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	ToolID         uuid.UUID `gorm:"type:uuid;not null"`
	FightSessionID uuid.UUID `gorm:"type:uuid;not null"`
	PointsGained   int32     `gorm:"not null"`
	LevelAfter     int16     `gorm:"not null"`
	CreatedAt      time.Time
}

func (ToolMasteryEvent) TableName() string {
	return "tool_mastery_events"
}
