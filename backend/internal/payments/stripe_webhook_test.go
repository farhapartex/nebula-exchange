package payments_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stripe/stripe-go/v84/webhook"

	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/ledger/ledgertest"
	"nebula-exchange/backend/internal/payments"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/users"
)

const testWebhookSecret = "whsec_test_secret_for_unit_tests"

func checkoutEvent(eventID, eventType, paymentID string, amountCents int64, paymentStatus string) []byte {
	encoded, _ := json.Marshal(map[string]any{
		"id":          eventID,
		"object":      "event",
		"type":        eventType,
		"api_version": "2025-01-01",
		"data": map[string]any{"object": map[string]any{
			"id":             "cs_test_" + eventID,
			"object":         "checkout.session",
			"amount_total":   amountCents,
			"currency":       "usd",
			"payment_status": paymentStatus,
			"metadata":       map[string]string{"payment_id": paymentID},
		}},
	})
	return encoded
}

func TestStripeWebhookSettlesOnceAndExpires(t *testing.T) {
	harness := newPaymentsHarness(t)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	webhookRouter := httpserver.NewRouter(httpserver.RouterOptions{Logger: testLogger},
		payments.NewStripeWebhookHandler(harness.pool, harness.settler, testWebhookSecret, testLogger, time.Now))

	deliver := func(payload []byte, secret string) int {
		signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: payload, Secret: secret})
		webhookRequest := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/stripe", bytes.NewReader(payload))
		webhookRequest.Header.Set("Stripe-Signature", signed.Header)
		recorder := httptest.NewRecorder()
		webhookRouter.ServeHTTP(recorder, webhookRequest)
		return recorder.Code
	}

	player := harness.createPlayer(t, users.StatusPendingPayment)
	entryFee := harness.createPayment(t, player, `{"purpose":"ENTRY_FEE","method":"card"}`)
	completed := checkoutEvent("evt_completed", "checkout.session.completed", entryFee.ID.String(), 500, "paid")

	if status := deliver(completed, "whsec_wrong"); status != http.StatusBadRequest {
		t.Fatalf("a forged signature must be rejected, got %d", status)
	}
	if status := deliver(completed, testWebhookSecret); status != http.StatusOK {
		t.Fatalf("completed event status %d", status)
	}
	if status := deliver(completed, testWebhookSecret); status != http.StatusOK {
		t.Fatalf("replayed event status %d", status)
	}
	if accountStatus(t, harness.pool, player) != "ACTIVE" {
		t.Fatal("the entry fee webhook must activate the account")
	}
	if card := ledgertest.BalanceOf(t, harness.pool, ledger.PlayerNC(player, ledger.BucketCard)); card.Available != 2_000_000 {
		t.Fatalf("a replayed webhook must not credit twice, card %d", card.Available)
	}

	topupPlayer := harness.createPlayer(t, users.StatusActive)
	topup := harness.createPayment(t, topupPlayer, `{"purpose":"TOPUP","method":"card","amount_nc":"10"}`)
	if status := deliver(checkoutEvent("evt_expired", "checkout.session.expired", topup.ID.String(), 1000, "unpaid"), testWebhookSecret); status != http.StatusOK {
		t.Fatalf("expired event status %d", status)
	}
	var topupStatus string
	harness.pool.QueryRow(context.Background(), "SELECT status FROM payments WHERE id = $1", topup.ID).Scan(&topupStatus)
	if topupStatus != "EXPIRED" {
		t.Fatalf("expired session must expire the payment, got %s", topupStatus)
	}

	if status := deliver(checkoutEvent("evt_unpaid", "checkout.session.completed", topup.ID.String(), 1000, "unpaid"), testWebhookSecret); status != http.StatusOK {
		t.Fatalf("unpaid completion status %d", status)
	}
	if status := deliver(checkoutEvent("evt_foreign", "checkout.session.completed", "not-our-payment", 1000, "paid"), testWebhookSecret); status != http.StatusOK {
		t.Fatalf("foreign session status %d", status)
	}
	if status := deliver([]byte(`{"id":"evt_other","object":"event","type":"customer.created","api_version":"2025-01-01","data":{"object":{}}}`), testWebhookSecret); status != http.StatusOK {
		t.Fatalf("ignored event type status %d", status)
	}
	ledgertest.RequireIntegrity(t, harness.pool)
}

func TestCancellingACheckoutFailsThePendingPayment(t *testing.T) {
	harness := newPaymentsHarness(t)
	player := harness.createPlayer(t, users.StatusPendingPayment)
	entryFee := harness.createPayment(t, player, `{"purpose":"ENTRY_FEE","method":"card"}`)

	status, cancelled := harness.send(t, player, http.MethodPost, "/payments/"+entryFee.ID.String()+"/cancellations", "")
	cancelledPayment := cancelled["data"].(map[string]any)
	if status != http.StatusOK || cancelledPayment["status"] != "FAILED" || cancelledPayment["purpose_failure_code"] != "CHECKOUT_CANCELLED" {
		t.Fatalf("cancel: %d %v", status, cancelled)
	}
	if outcome := harness.settle(t, entryFee, "evt_after_cancel"); outcome.WasSettled {
		t.Fatal("a cancelled payment must not settle")
	}

	settled := harness.createPayment(t, player, `{"purpose":"ENTRY_FEE","method":"card"}`)
	harness.settle(t, settled, "evt_settled")
	if _, again := harness.send(t, player, http.MethodPost, "/payments/"+settled.ID.String()+"/cancellations", ""); again["data"].(map[string]any)["status"] != "SUCCEEDED" {
		t.Fatalf("cancelling a paid payment must change nothing: %v", again)
	}
	otherPlayer := harness.createPlayer(t, users.StatusActive)
	if status, _ := harness.send(t, otherPlayer, http.MethodPost, "/payments/"+entryFee.ID.String()+"/cancellations", ""); status != http.StatusNotFound {
		t.Fatalf("someone else's payment: %d", status)
	}
}
