package models

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusOpen     PaymentStatus = "OPEN"
	PaymentStatusPaid     PaymentStatus = "PAID"
	PaymentStatusExpired  PaymentStatus = "EXPIRED"
	PaymentStatusFailed   PaymentStatus = "FAILED"
	PaymentStatusRefunded PaymentStatus = "REFUNDED"
	PaymentStatusDisputed PaymentStatus = "DISPUTED"
)

type PaymentMethod string

const (
	PaymentMethodCard   PaymentMethod = "CARD"
	PaymentMethodWallet PaymentMethod = "WALLET"
)

type WalletPaymentAsset string

const (
	WalletPaymentAssetETH  WalletPaymentAsset = "ETH"
	WalletPaymentAssetUSDC WalletPaymentAsset = "USDC"
)

var SettledPaymentStatuses = []PaymentStatus{PaymentStatusPaid, PaymentStatusRefunded, PaymentStatusDisputed}

func (status PaymentStatus) WasPaid() bool {
	for _, settledStatus := range SettledPaymentStatuses {
		if status == settledStatus {
			return true
		}
	}
	return false
}

type Payment struct {
	ID                      uuid.UUID     `gorm:"type:uuid;primaryKey"`
	UserID                  uuid.UUID     `gorm:"type:uuid;not null"`
	PlanID                  string        `gorm:"not null"`
	Status                  PaymentStatus `gorm:"type:payment_status;not null"`
	PaymentMethod           PaymentMethod `gorm:"type:payment_method;not null"`
	Currency                string        `gorm:"not null"`
	SubtotalCents           int64         `gorm:"not null"`
	DiscountPercent         int           `gorm:"not null"`
	DiscountCents           int64         `gorm:"not null"`
	TotalCents              int64         `gorm:"not null"`
	StripeCheckoutSessionID *string
	StripePaymentIntentID   *string
	CheckoutURL             *string
	PaymentReference        *string
	PayerAddress            *string
	AuthorizationSignature  *string
	ReportedTransactionHash *string
	TransactionHash         *string
	BlockNumber             *int64
	WalletAsset             *WalletPaymentAsset `gorm:"type:wallet_payment_asset"`
	WalletAmountUnits       *string             `gorm:"type:numeric(78,0)"`
	CheckoutExpiresAt       time.Time           `gorm:"not null"`
	PaidAt                  *time.Time
	RefundedAt              *time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time
	Plan                    Plan             `gorm:"foreignKey:PlanID"`
	Chapters                []PaymentChapter `gorm:"foreignKey:PaymentID"`
}

func (Payment) TableName() string {
	return "payments"
}

func (payment Payment) ChapterIDs() []string {
	chapterIDs := make([]string, 0, len(payment.Chapters))
	for _, chapter := range payment.Chapters {
		chapterIDs = append(chapterIDs, chapter.ChapterID)
	}
	return chapterIDs
}

type PaymentChapter struct {
	PaymentID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	ChapterID     string    `gorm:"primaryKey"`
	ChapterNumber int       `gorm:"not null"`
	ChapterTitle  string    `gorm:"not null"`
	PriceCents    int64     `gorm:"not null"`
}

func (PaymentChapter) TableName() string {
	return "payment_chapters"
}
