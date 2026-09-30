package payments

import (
	"time"

	"github.com/google/uuid"

	"nebula-exchange/backend/internal/payments/paymentsstore"
	"nebula-exchange/backend/internal/platform/money"
	"nebula-exchange/backend/internal/purpose"
)

type Method string

const (
	MethodCard   Method = "card"
	MethodCrypto Method = "crypto"
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusSucceeded Status = "SUCCEEDED"
	StatusFailed    Status = "FAILED"
	StatusExpired   Status = "EXPIRED"
)

type Payment struct {
	ID                 uuid.UUID    `json:"id"`
	Purpose            purpose.Kind `json:"purpose"`
	Method             Method       `json:"method"`
	Status             Status       `json:"status"`
	Amount             money.Micro  `json:"amount"`
	Credited           *money.Micro `json:"credited"`
	SKU                *string      `json:"sku"`
	UpgradeID          *string      `json:"upgrade_id"`
	Quantity           int          `json:"quantity"`
	PurposeStatus      string       `json:"purpose_status"`
	PurposeFailureCode *string      `json:"purpose_failure_code"`
	CheckoutURL        *string      `json:"checkout_url"`
	ExpiresAt          time.Time    `json:"expires_at"`
	SucceededAt        *time.Time   `json:"succeeded_at"`
	CreatedAt          time.Time    `json:"created_at"`
}

func paymentFromRow(paymentRow paymentsstore.Payment) Payment {
	payment := Payment{
		ID:                 paymentRow.ID,
		Purpose:            purpose.Kind(paymentRow.Purpose),
		Method:             Method(paymentRow.Method),
		Status:             Status(paymentRow.Status),
		Amount:             money.Micro(paymentRow.AmountMicro),
		SKU:                paymentRow.Sku,
		UpgradeID:          paymentRow.UpgradeID,
		Quantity:           int(paymentRow.Quantity),
		PurposeStatus:      paymentRow.PurposeStatus,
		PurposeFailureCode: paymentRow.PurposeFailureCode,
		CheckoutURL:        paymentRow.CheckoutUrl,
		ExpiresAt:          paymentRow.ExpiresAt,
		SucceededAt:        paymentRow.SucceededAt,
		CreatedAt:          paymentRow.CreatedAt,
	}
	if paymentRow.CreditedMicro != nil {
		credited := money.Micro(*paymentRow.CreditedMicro)
		payment.Credited = &credited
	}
	return payment
}
