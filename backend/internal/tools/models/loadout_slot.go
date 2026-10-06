package models

import (
	"time"

	"github.com/google/uuid"
)

type LoadoutSlotKind string

const (
	LoadoutSlotWeapon LoadoutSlotKind = "WEAPON"
	LoadoutSlotGuard  LoadoutSlotKind = "GUARD"
)

type LoadoutSlot struct {
	UserID    uuid.UUID       `gorm:"type:uuid;primaryKey"`
	Slot      LoadoutSlotKind `gorm:"type:loadout_slot;primaryKey"`
	ToolID    *uuid.UUID      `gorm:"type:uuid"`
	UpdatedAt time.Time
	Tool      *Tool `gorm:"foreignKey:ToolID"`
}

func (LoadoutSlot) TableName() string {
	return "loadout_slots"
}
