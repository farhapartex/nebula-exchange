package models

import "time"

type ChainSyncCursor struct {
	Name        string `gorm:"primaryKey"`
	ChainID     int64  `gorm:"not null"`
	BlockNumber int64  `gorm:"not null"`
	BlockHash   string `gorm:"not null"`
	UpdatedAt   time.Time
}

func (ChainSyncCursor) TableName() string {
	return "chain_sync_cursors"
}
