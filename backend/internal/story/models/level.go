package models

import (
	"time"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type LevelKind string

const (
	LevelKindStory    LevelKind = "STORY"
	LevelKindTraining LevelKind = "TRAINING"
	LevelKindBoss     LevelKind = "BOSS"
)

type Level struct {
	ID                   string `gorm:"primaryKey"`
	ChapterID            *string
	Number               *int
	Kind                 LevelKind             `gorm:"type:level_kind;not null"`
	Title                string                `gorm:"not null"`
	Teaser               string                `gorm:"not null"`
	ArenaID              string                `gorm:"not null"`
	TimeLimitSeconds     int                   `gorm:"not null"`
	Difficulty           database.JSONDocument `gorm:"type:jsonb;not null"`
	StarRules            database.JSONDocument `gorm:"type:jsonb;not null"`
	FirstClearCoins      int64                 `gorm:"not null"`
	FirstClearExperience int                   `gorm:"not null"`
	ReplayExperience     int                   `gorm:"not null"`
	IsFree               bool                  `gorm:"not null"`
	IsPublished          bool                  `gorm:"not null"`
	CreatedAt            time.Time
	UpdatedAt            time.Time

	Chapter *Chapter     `gorm:"foreignKey:ChapterID"`
	Arena   Arena        `gorm:"foreignKey:ArenaID"`
	Enemies []LevelEnemy `gorm:"foreignKey:LevelID"`
	Slides  []StorySlide `gorm:"foreignKey:LevelID"`
}

func (Level) TableName() string {
	return "levels"
}
