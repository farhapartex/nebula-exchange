package models

import (
	"time"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type ToolCategory string

const (
	ToolCategoryWeapon ToolCategory = "WEAPON"
	ToolCategoryGuard  ToolCategory = "GUARD"
)

type ToolRarity string

const (
	ToolRarityCommon    ToolRarity = "COMMON"
	ToolRarityUncommon  ToolRarity = "UNCOMMON"
	ToolRarityRare      ToolRarity = "RARE"
	ToolRarityEpic      ToolRarity = "EPIC"
	ToolRarityLegendary ToolRarity = "LEGENDARY"
)

type ToolType struct {
	ID                  string                `gorm:"primaryKey"`
	Name                string                `gorm:"not null"`
	Description         string                `gorm:"not null"`
	Category            ToolCategory          `gorm:"type:tool_category;not null"`
	Rarity              ToolRarity            `gorm:"type:tool_rarity;not null"`
	BaseStats           database.JSONDocument `gorm:"type:jsonb;not null"`
	MasteryCurve        database.JSONDocument `gorm:"type:jsonb;not null"`
	MaxMasteryLevel     int16                 `gorm:"not null"`
	ShopPriceCoins      *int64
	IsTradeable         bool `gorm:"not null"`
	MaxSupply           *int32
	ImageSVG            *string
	MinimumFighterLevel int16 `gorm:"not null"`
	IntroducedInLevelID *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (ToolType) TableName() string {
	return "tool_types"
}
