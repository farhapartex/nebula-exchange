package models

import (
	"time"

	"github.com/google/uuid"
)

type ChapterUnlockSource string

const (
	ChapterUnlockSourcePurchase ChapterUnlockSource = "PURCHASE"
	ChapterUnlockSourceGrant    ChapterUnlockSource = "GRANT"
)

type ChapterUnlock struct {
	UserID     uuid.UUID           `gorm:"type:uuid;primaryKey"`
	ChapterID  string              `gorm:"primaryKey"`
	Source     ChapterUnlockSource `gorm:"type:chapter_unlock_source;not null"`
	PaymentID  *uuid.UUID          `gorm:"type:uuid"`
	UnlockedAt time.Time           `gorm:"not null"`
	CreatedAt  time.Time
}

func (ChapterUnlock) TableName() string {
	return "chapter_unlocks"
}
