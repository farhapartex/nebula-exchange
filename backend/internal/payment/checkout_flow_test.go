package payment_test

import (
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestCheckoutIsPricedOnTheServerAndReturnsTheStripeLink(t *testing.T) {
	harness := newPaymentHarness(t)

	status, body := harness.startCheckout(playerToken, "chapter-bundle", 3)
	if status != http.StatusCreated {
		t.Fatalf("got %d %v, want 201", status, body)
	}
	paymentID := body.Data["id"].(string)
	if body.Data["checkout_url"] != "https://checkout.stripe.test/cs_test_1" || body.Data["expires_at"] == nil {
		t.Fatalf("unexpected checkout %v", body.Data)
	}
	stripeRequest := harness.gateway.createdRequests[0]
	if stripeRequest.AmountCents != 1347 || stripeRequest.Currency != "usd" || stripeRequest.PaymentID.String() != paymentID {
		t.Fatalf("stripe got the wrong amount %+v", stripeRequest)
	}
	if stripeRequest.SuccessURL != testFrontendURL+"/subscription?checkout="+paymentID || stripeRequest.CancelURL != testFrontendURL+"/fight?checkout=cancelled" {
		t.Fatalf("unexpected return links %s %s", stripeRequest.SuccessURL, stripeRequest.CancelURL)
	}
	if stripeRequest.ProductName != "Street Born: Chapters 2 to 4" || !strings.Contains(stripeRequest.ProductDescription, "10% off") {
		t.Fatalf("unexpected product %q %q", stripeRequest.ProductName, stripeRequest.ProductDescription)
	}
	var storedChapterIDs []string
	harness.database.Table("payment_chapters").Where("payment_id = ?", paymentID).Order("chapter_number").Pluck("chapter_id", &storedChapterIDs)
	if !reflect.DeepEqual(storedChapterIDs, []string{"2", "3", "4"}) || harness.paymentStatus(paymentID) != "OPEN" {
		t.Fatalf("payment stored %v with status %s", storedChapterIDs, harness.paymentStatus(paymentID))
	}

	if status, body := harness.getData("/checkout-sessions/"+paymentID, playerToken); status != http.StatusOK || body.Data["status"] != "OPEN" {
		t.Fatalf("got %d %v, want an open checkout", status, body)
	}
	if status, _ := harness.getData("/checkout-sessions/"+paymentID, otherPlayerToken); status != http.StatusNotFound {
		t.Fatalf("another player saw the checkout, got %d", status)
	}
}

func TestCheckoutRejectsPlansAndCountsThatAreNotOffered(t *testing.T) {
	harness := newPaymentHarness(t)

	if status, body := harness.startCheckout(playerToken, "lifetime-pass", 1); status != http.StatusUnprocessableEntity || body.Error.Details["plan_id"] == "" {
		t.Fatalf("got %d %v for an unknown plan", status, body.Error)
	}
	if status, body := harness.startCheckout(playerToken, "chapter-bundle", 4); status != http.StatusUnprocessableEntity || body.Error.Details["chapter_count"] == "" {
		t.Fatalf("got %d %v for a bundle of every chapter", status, body.Error)
	}
	if status, body := harness.startCheckout(playerToken, "single-chapter", 0); status != http.StatusUnprocessableEntity {
		t.Fatalf("got %d %v for zero chapters", status, body.Error)
	}
	if len(harness.gateway.createdRequests) != 0 {
		t.Fatal("stripe must not be called for a rejected checkout")
	}
}

func TestANewCheckoutExpiresTheOlderUnpaidOne(t *testing.T) {
	harness := newPaymentHarness(t)

	_, firstCheckout := harness.startCheckout(playerToken, "single-chapter", 1)
	status, secondCheckout := harness.startCheckout(playerToken, "all-chapters", 4)
	if status != http.StatusCreated {
		t.Fatalf("got %d %v", status, secondCheckout)
	}
	if !reflect.DeepEqual(harness.gateway.expiredSessions, []string{"cs_test_1"}) {
		t.Fatalf("expected the first stripe session to be expired, got %v", harness.gateway.expiredSessions)
	}
	firstPaymentID := firstCheckout.Data["id"].(string)
	if harness.paymentStatus(firstPaymentID) != "EXPIRED" || harness.paymentStatus(secondCheckout.Data["id"].(string)) != "OPEN" {
		t.Fatal("only the newest checkout may stay open")
	}
	if _, body := harness.getData("/checkout-sessions/"+firstPaymentID, playerToken); body.Data["status"] != "EXPIRED" {
		t.Fatalf("got %v, want an expired checkout", body.Data)
	}
}

func TestAPaidWebhookUnlocksTheChaptersOnceAndListsTheSubscription(t *testing.T) {
	harness := newPaymentHarness(t)
	_, checkout := harness.startCheckout(playerToken, "chapter-bundle", 2)
	paymentID := checkout.Data["id"].(string)
	completedEvent := stripeEvent("evt_paid", "checkout.session.completed", completedCheckoutObject("cs_test_1", paymentID, "pi_paid", 948))

	if status := harness.sendStripeEvent(completedEvent, "whsec_wrong"); status != http.StatusBadRequest {
		t.Fatalf("got %d for a forged signature, want 400", status)
	}
	if owned := harness.ownedChapterNumbers(playerToken); len(owned) != 0 {
		t.Fatalf("a forged webhook unlocked %v", owned)
	}

	for attempt := 1; attempt <= 2; attempt++ {
		if status := harness.sendStripeEvent(completedEvent, testWebhookSecret); status != http.StatusOK {
			t.Fatalf("attempt %d got %d, want 200", attempt, status)
		}
	}
	if owned := harness.ownedChapterNumbers(playerToken); !reflect.DeepEqual(owned, []float64{2, 3}) {
		t.Fatalf("got owned chapters %v, want 2 and 3", owned)
	}
	var storedEvents int64
	harness.database.Table("stripe_webhook_events").Count(&storedEvents)
	if storedEvents != 1 || harness.paymentStatus(paymentID) != "PAID" {
		t.Fatalf("got %d stored events and status %s", storedEvents, harness.paymentStatus(paymentID))
	}
	if _, body := harness.getData("/checkout-sessions/"+paymentID, playerToken); body.Data["status"] != "PAID" {
		t.Fatalf("got %v, want a paid checkout", body.Data)
	}

	status, subscriptions := harness.get("/subscriptions", playerToken)
	if status != http.StatusOK || len(subscriptions.Data) != 1 {
		t.Fatalf("got %d %v", status, subscriptions)
	}
	subscription := subscriptions.Data[0]
	if subscription["plan_name"] != "A number of chapters" || subscription["status"] != "PAID" || subscription["total_cents"] != "948" || subscription["discount_percent"] != float64(5) || len(subscription["chapters"].([]any)) != 2 || subscription["refunded_at"] != nil {
		t.Fatalf("unexpected subscription %v", subscription)
	}
	if _, otherSubscriptions := harness.get("/subscriptions", otherPlayerToken); len(otherSubscriptions.Data) != 0 {
		t.Fatal("another player must not see this purchase")
	}
}

func TestAWebhookWithTheWrongAmountDoesNotUnlockChapters(t *testing.T) {
	harness := newPaymentHarness(t)
	_, checkout := harness.startCheckout(playerToken, "single-chapter", 1)
	paymentID := checkout.Data["id"].(string)

	tamperedEvent := stripeEvent("evt_cheap", "checkout.session.completed", completedCheckoutObject("cs_test_1", paymentID, "pi_cheap", 1))
	if status := harness.sendStripeEvent(tamperedEvent, testWebhookSecret); status != http.StatusOK {
		t.Fatalf("got %d", status)
	}
	if owned := harness.ownedChapterNumbers(playerToken); len(owned) != 0 || harness.paymentStatus(paymentID) != "OPEN" {
		t.Fatalf("a 1 cent payment unlocked %v", owned)
	}
}

func TestARefundLocksTheChaptersAgainAndKeepsTheRecord(t *testing.T) {
	harness := newPaymentHarness(t)
	_, checkout := harness.startCheckout(playerToken, "chapter-bundle", 2)
	paymentID := checkout.Data["id"].(string)
	harness.sendStripeEvent(stripeEvent("evt_paid", "checkout.session.completed", completedCheckoutObject("cs_test_1", paymentID, "pi_refund_me", 948)), testWebhookSecret)

	partialRefund := stripeEvent("evt_partial", "charge.refunded", map[string]any{"object": "charge", "payment_intent": "pi_refund_me", "refunded": false})
	harness.sendStripeEvent(partialRefund, testWebhookSecret)
	if owned := harness.ownedChapterNumbers(playerToken); len(owned) != 2 {
		t.Fatalf("a partial refund changed the chapters to %v", owned)
	}

	fullRefund := stripeEvent("evt_full", "charge.refunded", map[string]any{"object": "charge", "payment_intent": "pi_refund_me", "refunded": true})
	if status := harness.sendStripeEvent(fullRefund, testWebhookSecret); status != http.StatusOK {
		t.Fatalf("got %d", status)
	}
	if owned := harness.ownedChapterNumbers(playerToken); len(owned) != 0 {
		t.Fatalf("refunded chapters are still unlocked: %v", owned)
	}
	_, subscriptions := harness.get("/subscriptions", playerToken)
	if len(subscriptions.Data) != 1 || subscriptions.Data[0]["status"] != "REFUNDED" || subscriptions.Data[0]["refunded_at"] == nil {
		t.Fatalf("unexpected subscriptions %v", subscriptions.Data)
	}
}

func TestADisputeLocksOnlyChaptersNoOtherPaymentCovers(t *testing.T) {
	harness := newPaymentHarness(t)
	_, firstCheckout := harness.startCheckout(playerToken, "single-chapter", 1)
	firstPaymentID := firstCheckout.Data["id"].(string)
	harness.sendStripeEvent(stripeEvent("evt_first", "checkout.session.completed", completedCheckoutObject("cs_test_1", firstPaymentID, "pi_first", 499)), testWebhookSecret)

	_, secondCheckout := harness.startCheckout(playerToken, "single-chapter", 1)
	secondPaymentID := secondCheckout.Data["id"].(string)
	harness.sendStripeEvent(stripeEvent("evt_second", "checkout.session.completed", completedCheckoutObject("cs_test_2", secondPaymentID, "pi_second", 499)), testWebhookSecret)
	if owned := harness.ownedChapterNumbers(playerToken); !reflect.DeepEqual(owned, []float64{2, 3}) {
		t.Fatalf("got %v, want chapters 2 and 3", owned)
	}

	harness.sendStripeEvent(stripeEvent("evt_dispute", "charge.dispute.created", map[string]any{"object": "dispute", "payment_intent": "pi_second"}), testWebhookSecret)
	if owned := harness.ownedChapterNumbers(playerToken); !reflect.DeepEqual(owned, []float64{2}) {
		t.Fatalf("got %v after the dispute, want only chapter 2", owned)
	}
	if harness.paymentStatus(secondPaymentID) != "DISPUTED" || harness.paymentStatus(firstPaymentID) != "PAID" {
		t.Fatal("only the disputed payment may change")
	}
}

func TestCheckingAnOpenCheckoutPicksUpAPaymentStripeAlreadyTook(t *testing.T) {
	harness := newPaymentHarness(t)
	_, checkout := harness.startCheckout(playerToken, "single-chapter", 1)
	paymentID := checkout.Data["id"].(string)

	harness.gateway.payInStripe("cs_test_1", "pi_polled")
	if _, body := harness.getData("/checkout-sessions/"+paymentID, playerToken); body.Data["status"] != "PAID" {
		t.Fatalf("got %v, want the checkout to be paid", body.Data)
	}
	if owned := harness.ownedChapterNumbers(playerToken); !reflect.DeepEqual(owned, []float64{2}) {
		t.Fatalf("got %v, want chapter 2", owned)
	}

	harness.sendStripeEvent(stripeEvent("evt_late", "checkout.session.completed", completedCheckoutObject("cs_test_1", paymentID, "pi_polled", 499)), testWebhookSecret)
	var unlockCount int64
	harness.database.Table("chapter_unlocks").Where("user_id = ?", harness.playerID).Count(&unlockCount)
	if unlockCount != 1 {
		t.Fatalf("the late webhook unlocked again, got %d unlocks", unlockCount)
	}
}

func TestAPreviousCheckoutThatWasPaidIsNotChargedAgain(t *testing.T) {
	harness := newPaymentHarness(t)
	_, checkout := harness.startCheckout(playerToken, "single-chapter", 1)
	harness.gateway.payInStripe("cs_test_1", "pi_first")

	if status, body := harness.startCheckout(playerToken, "single-chapter", 1); status != http.StatusConflict {
		t.Fatalf("got %d %v, want 409", status, body)
	}
	if harness.paymentStatus(checkout.Data["id"].(string)) != "PAID" || len(harness.gateway.createdRequests) != 1 {
		t.Fatal("the paid checkout must be settled and no new checkout created")
	}
}

func TestCheckoutFailsCleanlyWhenStripeIsDown(t *testing.T) {
	harness := newPaymentHarness(t)
	harness.gateway.createErr = errors.New("stripe is down")

	if status, _ := harness.startCheckout(playerToken, "single-chapter", 1); status != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want 503", status)
	}
	var failedPayments int64
	harness.database.Table("payments").Where("status = ?", "FAILED").Count(&failedPayments)
	harness.gateway.createErr = nil
	if status, _ := harness.startCheckout(playerToken, "single-chapter", 1); failedPayments != 1 || status != http.StatusCreated {
		t.Fatalf("got %d failed payments and status %d on retry", failedPayments, status)
	}
}

func TestPaymentsAreOffWithoutStripeKeys(t *testing.T) {
	harness := newPaymentHarness(t, harnessOptions{withoutStripe: true})

	if status, _ := harness.startCheckout(playerToken, "single-chapter", 1); status != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want 503", status)
	}
	if status := harness.sendStripeEvent(stripeEvent("evt_any", "checkout.session.completed", map[string]any{}), testWebhookSecret); status != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want 503", status)
	}
	if status, _ := harness.get("/plans", playerToken); status != http.StatusOK {
		t.Fatalf("plans must still load, got %d", status)
	}
}
