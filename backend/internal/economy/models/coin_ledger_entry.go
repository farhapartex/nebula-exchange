package models

import (
	"time"

	"github.com/google/uuid"
)

type CoinEntryReason string

const (
	CoinEntryReasonWelcomeGift      CoinEntryReason = "WELCOME_GIFT"
	CoinEntryReasonCoinPack         CoinEntryReason = "COIN_PACK"
	CoinEntryReasonCoinPackClawback CoinEntryReason = "COIN_PACK_CLAWBACK"
)

type CoinLedgerEntry struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey"`
	UserID       uuid.UUID       `gorm:"type:uuid;not null"`
	Amount       int64           `gorm:"not null"`
	BalanceAfter int64           `gorm:"not null"`
	Reason       CoinEntryReason `gorm:"type:coin_entry_reason;not null"`
	PaymentID    *uuid.UUID      `gorm:"type:uuid"`
	CreatedAt    time.Time       `gorm:"not null"`
}

func (CoinLedgerEntry) TableName() string {
	return "coin_ledger_entries"
}
