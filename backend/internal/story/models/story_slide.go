package models

import (
	"time"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/google/uuid"
)

type SlideKind string

const (
	SlideKindSlide        SlideKind = "SLIDE"
	SlideKindCallToAction SlideKind = "CALL_TO_ACTION"
)

type StorySlide struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	LevelID     string    `gorm:"not null"`
	Position    int       `gorm:"not null"`
	Kind        SlideKind `gorm:"type:slide_kind;not null"`
	Eyebrow     *string
	Heading     string `gorm:"not null"`
	Body        string `gorm:"not null"`
	ImageKey    *string
	Palette     database.JSONDocument `gorm:"type:jsonb;not null"`
	ButtonLabel *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (StorySlide) TableName() string {
	return "story_slides"
}
