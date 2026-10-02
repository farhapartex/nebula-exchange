package payments

import (
	"context"
	"errors"
	"net/url"

	"github.com/google/uuid"
)

var (
	ErrCardPaymentsUnavailable = errors.New("card payments are not configured")
	ErrCheckoutAlreadyPaid     = errors.New("checkout session was already paid")
)

type CheckoutSession struct {
	ProviderSessionID string
	URL               string
}

type CheckoutRequest struct {
	PaymentID   uuid.UUID
	Description string
	AmountCents int64
}

type CardCheckout interface {
	CreateSession(ctx context.Context, request CheckoutRequest) (CheckoutSession, error)
	CancelSession(ctx context.Context, providerSessionID string) error
}

type DevelopmentCheckout struct {
	frontendBaseURL string
}

func NewDevelopmentCheckout(frontendBaseURL string) *DevelopmentCheckout {
	return &DevelopmentCheckout{frontendBaseURL: frontendBaseURL}
}

func (checkout *DevelopmentCheckout) CreateSession(_ context.Context, request CheckoutRequest) (CheckoutSession, error) {
	resultURL := checkout.frontendBaseURL + "/payment/result?" + url.Values{"id": {request.PaymentID.String()}}.Encode()
	return CheckoutSession{ProviderSessionID: "development_" + request.PaymentID.String(), URL: resultURL}, nil
}

func (checkout *DevelopmentCheckout) CancelSession(context.Context, string) error {
	return nil
}

type UnavailableCheckout struct{}

func (UnavailableCheckout) CancelSession(context.Context, string) error {
	return ErrCardPaymentsUnavailable
}

func (UnavailableCheckout) CreateSession(context.Context, CheckoutRequest) (CheckoutSession, error) {
	return CheckoutSession{}, ErrCardPaymentsUnavailable
}
