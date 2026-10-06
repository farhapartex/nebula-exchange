package models

import (
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type FightStatus string

const (
	FightStatusStarted   FightStatus = "STARTED"
	FightStatusFinished  FightStatus = "FINISHED"
	FightStatusAbandoned FightStatus = "ABANDONED"
	FightStatusRejected  FightStatus = "REJECTED"
)

type FightOutcome string

const (
	FightOutcomeWon  FightOutcome = "WON"
	FightOutcomeLost FightOutcome = "LOST"
	FightOutcomeDraw FightOutcome = "DRAW"
)

type FightSession struct {
	ID               uuid.UUID             `gorm:"type:uuid;primaryKey"`
	UserID           uuid.UUID             `gorm:"type:uuid;not null"`
	LevelID          string                `gorm:"not null"`
	Status           FightStatus           `gorm:"type:fight_status;not null"`
	Seed             int64                 `gorm:"not null"`
	Loadout          database.JSONDocument `gorm:"type:jsonb;not null"`
	Outcome          *FightOutcome         `gorm:"type:fight_outcome"`
	WavesCleared     *int
	Stars            *int
	DurationMS       *int `gorm:"column:duration_ms"`
	DamageDealt      *int
	DamageTaken      *int
	MovesUsed        database.JSONDocument `gorm:"type:jsonb"`
	InputLog         database.JSONDocument `gorm:"type:jsonb"`
	RewardCoins      *int64
	RewardExperience *int
	StartedAt        time.Time `gorm:"not null"`
	FinishedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (FightSession) TableName() string {
	return "fight_sessions"
}
