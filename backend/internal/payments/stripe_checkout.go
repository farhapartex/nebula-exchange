package payments

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/stripe/stripe-go/v84"
)

const (
	stripeSessionLifetime = 24*time.Hour - time.Minute
	stripeCurrency        = "usd"
)

type StripeCheckout struct {
	client          *stripe.Client
	frontendBaseURL string
	now             func() time.Time
}

func NewStripeCheckout(secretKey, frontendBaseURL string, now func() time.Time) *StripeCheckout {
	return &StripeCheckout{client: stripe.NewClient(secretKey), frontendBaseURL: frontendBaseURL, now: now}
}

func (checkout *StripeCheckout) CreateSession(ctx context.Context, request CheckoutRequest) (CheckoutSession, error) {
	paymentID := request.PaymentID.String()
	paymentQuery := url.Values{"id": {paymentID}}.Encode()
	sessionParams := &stripe.CheckoutSessionCreateParams{
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		ClientReferenceID:  stripe.String(paymentID),
		SuccessURL:         stripe.String(checkout.frontendBaseURL + "/payment/result?" + paymentQuery),
		CancelURL:          stripe.String(checkout.frontendBaseURL + "/payment/cancelled?" + paymentQuery),
		ExpiresAt:          stripe.Int64(checkout.now().Add(stripeSessionLifetime).Unix()),
		Metadata:           map[string]string{"payment_id": paymentID},
		PaymentIntentData: &stripe.CheckoutSessionCreatePaymentIntentDataParams{
			Metadata: map[string]string{"payment_id": paymentID},
		},
		LineItems: []*stripe.CheckoutSessionCreateLineItemParams{{
			Quantity: stripe.Int64(1),
			PriceData: &stripe.CheckoutSessionCreateLineItemPriceDataParams{
				Currency:   stripe.String(stripeCurrency),
				UnitAmount: stripe.Int64(request.AmountCents),
				ProductData: &stripe.CheckoutSessionCreateLineItemPriceDataProductDataParams{
					Name: stripe.String(request.Description),
				},
			},
		}},
	}
	sessionParams.SetIdempotencyKey("checkout-" + paymentID)

	session, err := checkout.client.V1CheckoutSessions.Create(ctx, sessionParams)
	if err != nil {
		return CheckoutSession{}, fmt.Errorf("create stripe checkout session: %w", err)
	}
	return CheckoutSession{ProviderSessionID: session.ID, URL: session.URL}, nil
}

func (checkout *StripeCheckout) CancelSession(ctx context.Context, providerSessionID string) error {
	session, err := checkout.client.V1CheckoutSessions.Retrieve(ctx, providerSessionID, &stripe.CheckoutSessionRetrieveParams{})
	if err != nil {
		return fmt.Errorf("check stripe checkout session: %w", err)
	}
	switch session.Status {
	case stripe.CheckoutSessionStatusComplete:
		return ErrCheckoutAlreadyPaid
	case stripe.CheckoutSessionStatusExpired:
		return nil
	}
	if _, err := checkout.client.V1CheckoutSessions.Expire(ctx, providerSessionID, &stripe.CheckoutSessionExpireParams{}); err != nil {
		return fmt.Errorf("expire stripe checkout session: %w", err)
	}
	return nil
}
