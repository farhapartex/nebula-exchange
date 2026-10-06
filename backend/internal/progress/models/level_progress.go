package models

import (
	"time"

	"github.com/google/uuid"
)

type ProgressStatus string

const (
	ProgressStatusStarted   ProgressStatus = "STARTED"
	ProgressStatusCompleted ProgressStatus = "COMPLETED"
)

type LevelProgress struct {
	UserID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	LevelID          string         `gorm:"primaryKey"`
	Status           ProgressStatus `gorm:"type:progress_status;not null"`
	BestStars        *int
	Attempts         int       `gorm:"not null"`
	FirstStartedAt   time.Time `gorm:"not null"`
	FirstCompletedAt *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (LevelProgress) TableName() string {
	return "level_progress"
}
