package models

import (
	"time"

	"github.com/google/uuid"
)

type ToolStatus string

const (
	ToolStatusOwned  ToolStatus = "OWNED"
	ToolStatusListed ToolStatus = "LISTED"
)

type ToolSource string

const (
	ToolSourceShop   ToolSource = "SHOP"
	ToolSourceReward ToolSource = "REWARD"
	ToolSourceMarket ToolSource = "MARKET"
)

type Tool struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ToolTypeID    string     `gorm:"not null"`
	OwnerUserID   uuid.UUID  `gorm:"type:uuid;not null"`
	MasteryLevel  int16      `gorm:"not null"`
	MasteryPoints int32      `gorm:"not null"`
	FightsUsed    int32      `gorm:"not null"`
	HitsLanded    int32      `gorm:"not null"`
	Status        ToolStatus `gorm:"type:tool_status;not null"`
	AcquiredFrom  ToolSource `gorm:"type:tool_source;not null"`
	AcquiredAt    time.Time  `gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ToolType      ToolType `gorm:"foreignKey:ToolTypeID"`
}

func (Tool) TableName() string {
	return "tools"
}
