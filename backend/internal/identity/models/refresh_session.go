package models

import (
	"time"

	"github.com/google/uuid"
)

type RefreshSession struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null"`
	FamilyID     uuid.UUID  `gorm:"type:uuid;not null"`
	TokenHash    []byte     `gorm:"not null"`
	ReplacedByID *uuid.UUID `gorm:"type:uuid"`
	UserAgent    string     `gorm:"not null"`
	IPAddress    *string    `gorm:"type:inet"`
	ExpiresAt    time.Time  `gorm:"not null"`
	RevokedAt    *time.Time
	CreatedAt    time.Time
}

func (RefreshSession) TableName() string {
	return "refresh_sessions"
}
