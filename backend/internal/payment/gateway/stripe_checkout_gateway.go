package gateway

import (
	"context"

	"github.com/stripe/stripe-go/v87"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/service"
)

const (
	stripeCheckoutModePayment  = "payment"
	stripeCardPaymentMethod    = "card"
	stripePaymentIDMetadataKey = "payment_id"
	singleQuantity             = 1
)

type StripeCheckoutGateway struct {
	client *stripe.Client
}

func NewStripeCheckoutGateway(secretKey string) *StripeCheckoutGateway {
	return &StripeCheckoutGateway{client: stripe.NewClient(secretKey)}
}

func (gateway *StripeCheckoutGateway) CreateCheckout(ctx context.Context, checkoutRequest service.CheckoutRequest) (service.CreatedCheckout, error) {
	paymentID := checkoutRequest.PaymentID.String()
	paymentMetadata := map[string]string{stripePaymentIDMetadataKey: paymentID}
	checkoutSession, err := gateway.client.V1CheckoutSessions.Create(ctx, &stripe.CheckoutSessionCreateParams{
		Mode:                      stripe.String(stripeCheckoutModePayment),
		AllowedPaymentMethodTypes: stripe.StringSlice([]string{stripeCardPaymentMethod}),
		ClientReferenceID:         stripe.String(paymentID),
		Metadata:                  paymentMetadata,
		PaymentIntentData:         &stripe.CheckoutSessionCreatePaymentIntentDataParams{Metadata: paymentMetadata},
		SuccessURL:                stripe.String(checkoutRequest.SuccessURL),
		CancelURL:                 stripe.String(checkoutRequest.CancelURL),
		ExpiresAt:                 stripe.Int64(checkoutRequest.ExpiresAt.Unix()),
		LineItems: []*stripe.CheckoutSessionCreateLineItemParams{{
			Quantity: stripe.Int64(singleQuantity),
			PriceData: &stripe.CheckoutSessionCreateLineItemPriceDataParams{
				Currency:   stripe.String(checkoutRequest.Currency),
				UnitAmount: stripe.Int64(checkoutRequest.AmountCents),
				ProductData: &stripe.CheckoutSessionCreateLineItemPriceDataProductDataParams{
					Name:        stripe.String(checkoutRequest.ProductName),
					Description: stripe.String(checkoutRequest.ProductDescription),
				},
			},
		}},
	})
	if err != nil {
		return service.CreatedCheckout{}, err
	}
	return service.CreatedCheckout{SessionID: checkoutSession.ID, URL: checkoutSession.URL}, nil
}

func (gateway *StripeCheckoutGateway) ExpireCheckout(ctx context.Context, sessionID string) error {
	_, err := gateway.client.V1CheckoutSessions.Expire(ctx, sessionID, nil)
	return err
}

func (gateway *StripeCheckoutGateway) FetchCheckout(ctx context.Context, sessionID string) (service.CheckoutSnapshot, error) {
	checkoutSession, err := gateway.client.V1CheckoutSessions.Retrieve(ctx, sessionID, nil)
	if err != nil {
		return service.CheckoutSnapshot{}, err
	}
	var paymentIntentID string
	if checkoutSession.PaymentIntent != nil {
		paymentIntentID = checkoutSession.PaymentIntent.ID
	}
	return service.CheckoutSnapshot{
		SessionID:        checkoutSession.ID,
		IsPaid:           checkoutSession.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid,
		IsExpired:        checkoutSession.Status == stripe.CheckoutSessionStatusExpired,
		PaymentIntentID:  paymentIntentID,
		AmountTotalCents: checkoutSession.AmountTotal,
		Currency:         string(checkoutSession.Currency),
	}, nil
}
