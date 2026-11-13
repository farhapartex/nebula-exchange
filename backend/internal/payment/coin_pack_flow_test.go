package payment_test

import (
	"net/http"
	"reflect"
	"testing"
)

func TestCoinPacksAreListedInOrderWithAmountsAsStrings(t *testing.T) {
	harness := newPaymentHarness(t)

	status, body := harness.get("/coin-packs", playerToken)
	if status != http.StatusOK || len(body.Data) != 3 {
		t.Fatalf("got %d %v, want three coin packs", status, body.Data)
	}
	firstPack := body.Data[0]
	if firstPack["id"] != "coins-500" || firstPack["coins"] != "500" || firstPack["price_cents"] != "499" {
		t.Fatalf("unexpected first pack %v", firstPack)
	}
	if body.Data[2]["id"] != "coins-2400" {
		t.Fatalf("packs are out of order: %v", body.Data)
	}
	if status, _ := harness.get("/coin-packs", ""); status != http.StatusUnauthorized {
		t.Fatalf("got %d without a login, want 401", status)
	}
}

func TestCoinPackCheckoutChargesThePackPriceAndReturnsToTheSamePage(t *testing.T) {
	harness := newPaymentHarness(t)

	status, body := harness.startCoinCheckout(playerToken, map[string]any{"coin_pack_id": "coins-1100", "return_path": "/market"})
	if status != http.StatusCreated {
		t.Fatalf("got %d %v, want 201", status, body)
	}
	paymentID := body.Data["id"].(string)
	stripeRequest := harness.gateway.createdRequests[0]
	if stripeRequest.AmountCents != 999 || stripeRequest.ProductName != "Street Born: 1100 coins" {
		t.Fatalf("stripe got the wrong product %+v", stripeRequest)
	}
	if stripeRequest.SuccessURL != testFrontendURL+"/market?coin_checkout="+paymentID || stripeRequest.CancelURL != testFrontendURL+"/market?coin_checkout=cancelled" {
		t.Fatalf("unexpected return links %s %s", stripeRequest.SuccessURL, stripeRequest.CancelURL)
	}

	if status, _ := harness.startCoinCheckout(playerToken, map[string]any{"coin_pack_id": "coins-500", "return_path": "https://evil.test/steal"}); status != http.StatusCreated {
		t.Fatalf("got %d for an unknown return path", status)
	}
	if successURL := harness.gateway.createdRequests[1].SuccessURL; successURL[:len(testFrontendURL+"/fight?")] != testFrontendURL+"/fight?" {
		t.Fatalf("an unknown return path must fall back to the fight page, got %s", successURL)
	}
	if harness.paymentStatus(paymentID) != "EXPIRED" {
		t.Fatal("a new checkout must expire the older unpaid one")
	}
}

func TestCoinPackCheckoutRejectsUnknownPacksWalletsAndMixedRequests(t *testing.T) {
	harness := newPaymentHarness(t)

	if status, body := harness.startCoinCheckout(playerToken, map[string]any{"coin_pack_id": "coins-9999"}); status != http.StatusUnprocessableEntity || body.Error.Details["coin_pack_id"] == "" {
		t.Fatalf("got %d %v for an unknown pack", status, body.Error)
	}
	if status, body := harness.startCoinCheckout(playerToken, map[string]any{"coin_pack_id": "coins-500", "payment_method": "WALLET"}); status != http.StatusUnprocessableEntity || body.Error.Details["payment_method"] == "" {
		t.Fatalf("got %d %v for a wallet coin pack", status, body.Error)
	}
	if status, body := harness.startCoinCheckout(playerToken, map[string]any{"coin_pack_id": "coins-500", "plan_id": "single-chapter", "chapter_count": 1}); status != http.StatusUnprocessableEntity || body.Error.Details["plan_id"] == "" {
		t.Fatalf("got %d %v for a mixed request", status, body.Error)
	}
	if status, body := harness.startCoinCheckout(playerToken, map[string]any{}); status != http.StatusUnprocessableEntity || body.Error.Details["plan_id"] == "" {
		t.Fatalf("got %d %v for an empty request", status, body.Error)
	}
	if len(harness.gateway.createdRequests) != 0 {
		t.Fatal("stripe must not be called for a rejected checkout")
	}
}

func TestPaidCoinPackCreditsCoinsOnceAndStaysOutOfSubscriptions(t *testing.T) {
	harness := newPaymentHarness(t)
	_, checkout := harness.startCoinCheckout(playerToken, map[string]any{"coin_pack_id": "coins-500", "return_path": "/fight"})
	paymentID := checkout.Data["id"].(string)

	paidEvent := completedCheckoutObject("cs_test_1", paymentID, "pi_coins", 499)
	if status := harness.sendStripeEvent(stripeEvent("evt_coins_paid", "checkout.session.completed", paidEvent), testWebhookSecret); status != http.StatusOK {
		t.Fatalf("got %d", status)
	}
	harness.sendStripeEvent(stripeEvent("evt_coins_paid_again", "checkout.session.completed", paidEvent), testWebhookSecret)

	if balance := harness.coinBalance(harness.playerID); balance != 500 {
		t.Fatalf("got %d coins, want 500", balance)
	}
	if amounts := harness.coinLedgerAmounts(harness.playerID); !reflect.DeepEqual(amounts, []int64{500}) {
		t.Fatalf("got ledger %v, want one entry of 500", amounts)
	}
	if _, body := harness.getData("/checkout-sessions/"+paymentID, playerToken); body.Data["status"] != "PAID" {
		t.Fatalf("got %v, want a paid checkout", body.Data)
	}
	if _, subscriptions := harness.get("/subscriptions", playerToken); len(subscriptions.Data) != 0 {
		t.Fatalf("coin packs must not show as chapter purchases, got %v", subscriptions.Data)
	}
	if owned := harness.ownedChapterNumbers(playerToken); len(owned) != 0 {
		t.Fatalf("a coin pack unlocked chapters %v", owned)
	}
}

func TestRefundedCoinPackTakesBackUnspentCoins(t *testing.T) {
	harness := newPaymentHarness(t)
	_, checkout := harness.startCoinCheckout(playerToken, map[string]any{"coin_pack_id": "coins-500"})
	paymentID := checkout.Data["id"].(string)
	harness.sendStripeEvent(stripeEvent("evt_paid", "checkout.session.completed", completedCheckoutObject("cs_test_1", paymentID, "pi_refund_coins", 499)), testWebhookSecret)

	harness.sendStripeEvent(stripeEvent("evt_refund", "charge.refunded", map[string]any{"object": "charge", "payment_intent": "pi_refund_coins", "refunded": true}), testWebhookSecret)

	if balance := harness.coinBalance(harness.playerID); balance != 0 {
		t.Fatalf("got %d coins after the refund, want 0", balance)
	}
	if amounts := harness.coinLedgerAmounts(harness.playerID); !reflect.DeepEqual(amounts, []int64{500, -500}) {
		t.Fatalf("got ledger %v", amounts)
	}
	var flags int64
	harness.database.Table("coin_clawback_flags").Count(&flags)
	if flags != 0 || harness.paymentStatus(paymentID) != "REFUNDED" {
		t.Fatalf("got %d flags and status %s", flags, harness.paymentStatus(paymentID))
	}
}

func TestDisputedCoinPackWithSpentCoinsIsFlaggedForReview(t *testing.T) {
	harness := newPaymentHarness(t)
	_, checkout := harness.startCoinCheckout(playerToken, map[string]any{"coin_pack_id": "coins-500"})
	paymentID := checkout.Data["id"].(string)
	harness.sendStripeEvent(stripeEvent("evt_paid", "checkout.session.completed", completedCheckoutObject("cs_test_1", paymentID, "pi_dispute_coins", 499)), testWebhookSecret)
	if err := harness.database.Table("coin_balances").Where("user_id = ?", harness.playerID).Update("coins", 200).Error; err != nil {
		t.Fatalf("spend coins: %v", err)
	}

	disputeEvent := map[string]any{"object": "dispute", "payment_intent": "pi_dispute_coins"}
	harness.sendStripeEvent(stripeEvent("evt_dispute", "charge.dispute.created", disputeEvent), testWebhookSecret)
	harness.sendStripeEvent(stripeEvent("evt_dispute_again", "charge.dispute.created", disputeEvent), testWebhookSecret)

	if balance := harness.coinBalance(harness.playerID); balance != 0 {
		t.Fatalf("got %d coins after the dispute, want 0", balance)
	}
	var flag struct {
		PackCoins       int64
		ClawedBackCoins int64
		MissingCoins    int64
	}
	var flags int64
	harness.database.Table("coin_clawback_flags").Where("payment_id = ?", paymentID).Count(&flags)
	harness.database.Table("coin_clawback_flags").Where("payment_id = ?", paymentID).Take(&flag)
	if flags != 1 || flag.PackCoins != 500 || flag.ClawedBackCoins != 200 || flag.MissingCoins != 300 {
		t.Fatalf("got %d flags %+v, want one flag for 300 missing coins", flags, flag)
	}
}
