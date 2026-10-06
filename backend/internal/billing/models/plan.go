package models

import (
	"time"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type PlanKind string

const (
	PlanKindSingleChapter PlanKind = "SINGLE_CHAPTER"
	PlanKindChapterBundle PlanKind = "CHAPTER_BUNDLE"
	PlanKindAllChapters   PlanKind = "ALL_CHAPTERS"
)

type Plan struct {
	ID            string                `gorm:"primaryKey"`
	Kind          PlanKind              `gorm:"type:plan_kind;not null"`
	Name          string                `gorm:"not null"`
	Description   string                `gorm:"not null"`
	DiscountTiers database.JSONDocument `gorm:"type:jsonb;not null"`
	SortOrder     int                   `gorm:"not null"`
	IsActive      bool                  `gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (Plan) TableName() string {
	return "plans"
}
