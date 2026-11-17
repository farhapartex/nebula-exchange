package gateway_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stripe/stripe-go/v87/webhook"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/gateway"
)

const refundTestWebhookSecret = "whsec_refund_test"

func verifyChargeRefunded(t *testing.T, charge map[string]any) bool {
	t.Helper()
	payload, _ := json.Marshal(map[string]any{
		"id": "evt_refund", "object": "event", "type": "charge.refunded",
		"data": map[string]any{"object": charge},
	})
	signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: payload, Secret: refundTestWebhookSecret, Timestamp: time.Now()})
	event, err := gateway.NewStripeEventVerifier(refundTestWebhookSecret).VerifyEvent(payload, signedPayload.Header)
	if err != nil {
		t.Fatalf("verify event: %v", err)
	}
	if event.PaymentIntentID != "pi_refund" {
		t.Fatalf("got payment intent %q", event.PaymentIntentID)
	}
	return event.IsFullRefund
}

func TestOnlyAFullyRefundedChargeCountsAsAFullRefund(t *testing.T) {
	partialCharge := map[string]any{"object": "charge", "payment_intent": "pi_refund", "refunded": false, "amount_refunded": 100}
	if verifyChargeRefunded(t, partialCharge) {
		t.Fatal("a partial refund must not count as a full refund")
	}
	fullCharge := map[string]any{"object": "charge", "payment_intent": "pi_refund", "refunded": true, "amount_refunded": 499}
	if !verifyChargeRefunded(t, fullCharge) {
		t.Fatal("a full refund must count as a full refund")
	}
}
