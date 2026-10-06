package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CheckoutRequest struct {
	PaymentID          uuid.UUID
	ProductName        string
	ProductDescription string
	AmountCents        int64
	Currency           string
	SuccessURL         string
	CancelURL          string
	ExpiresAt          time.Time
}

type CreatedCheckout struct {
	SessionID string
	URL       string
}

type CheckoutSnapshot struct {
	SessionID        string
	IsPaid           bool
	IsExpired        bool
	PaymentIntentID  string
	AmountTotalCents int64
	Currency         string
}

type CheckoutGateway interface {
	CreateCheckout(ctx context.Context, checkoutRequest CheckoutRequest) (CreatedCheckout, error)
	ExpireCheckout(ctx context.Context, sessionID string) error
	FetchCheckout(ctx context.Context, sessionID string) (CheckoutSnapshot, error)
}

type PaymentEventKind string

const (
	PaymentEventCheckoutCompleted PaymentEventKind = "CHECKOUT_COMPLETED"
	PaymentEventCheckoutExpired   PaymentEventKind = "CHECKOUT_EXPIRED"
	PaymentEventChargeRefunded    PaymentEventKind = "CHARGE_REFUNDED"
	PaymentEventDisputeOpened     PaymentEventKind = "DISPUTE_OPENED"
	PaymentEventIgnored           PaymentEventKind = "IGNORED"
)

type PaymentEvent struct {
	ID              string
	Type            string
	Kind            PaymentEventKind
	Payload         []byte
	Checkout        CheckoutSnapshot
	PaymentID       string
	PaymentIntentID string
	IsFullRefund    bool
}

type PaymentEventVerifier interface {
	VerifyEvent(payload []byte, signatureHeader string) (PaymentEvent, error)
}
