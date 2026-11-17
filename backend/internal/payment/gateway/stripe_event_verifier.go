package gateway

import (
	"encoding/json"
	"fmt"

	"github.com/stripe/stripe-go/v87/webhook"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/service"
)

const (
	stripeCheckoutPaid    = "paid"
	stripeCheckoutExpired = "expired"
)

var paymentEventKindByStripeType = map[string]service.PaymentEventKind{
	"checkout.session.completed":               service.PaymentEventCheckoutCompleted,
	"checkout.session.async_payment_succeeded": service.PaymentEventCheckoutCompleted,
	"checkout.session.expired":                 service.PaymentEventCheckoutExpired,
	"charge.refunded":                          service.PaymentEventChargeRefunded,
	"charge.dispute.created":                   service.PaymentEventDisputeOpened,
}

type stripeEventEnvelope struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Object json.RawMessage `json:"object"`
	} `json:"data"`
}

type stripeCheckoutSessionObject struct {
	ID                string `json:"id"`
	ClientReferenceID string `json:"client_reference_id"`
	Status            string `json:"status"`
	PaymentStatus     string `json:"payment_status"`
	PaymentIntent     string `json:"payment_intent"`
	AmountTotal       int64  `json:"amount_total"`
	Currency          string `json:"currency"`
}

type stripeChargeObject struct {
	PaymentIntent string `json:"payment_intent"`
	Refunded      bool   `json:"refunded"`
}

type stripeDisputeObject struct {
	PaymentIntent string `json:"payment_intent"`
}

type StripeEventVerifier struct {
	webhookSecret string
}

func NewStripeEventVerifier(webhookSecret string) *StripeEventVerifier {
	return &StripeEventVerifier{webhookSecret: webhookSecret}
}

func (verifier *StripeEventVerifier) VerifyEvent(payload []byte, signatureHeader string) (service.PaymentEvent, error) {
	if err := webhook.ValidatePayload(payload, signatureHeader, verifier.webhookSecret); err != nil {
		return service.PaymentEvent{}, err
	}
	var envelope stripeEventEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return service.PaymentEvent{}, fmt.Errorf("stripe event: %w", err)
	}
	if envelope.ID == "" || envelope.Type == "" {
		return service.PaymentEvent{}, fmt.Errorf("stripe event has no id or type")
	}
	paymentEvent := service.PaymentEvent{ID: envelope.ID, Type: envelope.Type, Kind: service.PaymentEventIgnored, Payload: payload}
	eventKind, isHandledType := paymentEventKindByStripeType[envelope.Type]
	if !isHandledType {
		return paymentEvent, nil
	}
	paymentEvent.Kind = eventKind
	return paymentEvent, readEventObject(&paymentEvent, envelope.Data.Object)
}

func readEventObject(paymentEvent *service.PaymentEvent, object json.RawMessage) error {
	switch paymentEvent.Kind {
	case service.PaymentEventCheckoutCompleted, service.PaymentEventCheckoutExpired:
		var checkoutSession stripeCheckoutSessionObject
		if err := json.Unmarshal(object, &checkoutSession); err != nil {
			return fmt.Errorf("stripe checkout session: %w", err)
		}
		paymentEvent.PaymentID = checkoutSession.ClientReferenceID
		paymentEvent.PaymentIntentID = checkoutSession.PaymentIntent
		paymentEvent.Checkout = service.CheckoutSnapshot{
			SessionID:        checkoutSession.ID,
			IsPaid:           checkoutSession.PaymentStatus == stripeCheckoutPaid,
			IsExpired:        checkoutSession.Status == stripeCheckoutExpired,
			PaymentIntentID:  checkoutSession.PaymentIntent,
			AmountTotalCents: checkoutSession.AmountTotal,
			Currency:         checkoutSession.Currency,
		}
	case service.PaymentEventChargeRefunded:
		var charge stripeChargeObject
		if err := json.Unmarshal(object, &charge); err != nil {
			return fmt.Errorf("stripe charge: %w", err)
		}
		paymentEvent.PaymentIntentID = charge.PaymentIntent
		paymentEvent.IsFullRefund = charge.Refunded
	case service.PaymentEventDisputeOpened:
		var dispute stripeDisputeObject
		if err := json.Unmarshal(object, &dispute); err != nil {
			return fmt.Errorf("stripe dispute: %w", err)
		}
		paymentEvent.PaymentIntentID = dispute.PaymentIntent
	}
	return nil
}
