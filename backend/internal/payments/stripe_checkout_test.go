package payments_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"nebula-exchange/backend/internal/payments"
)

func TestStripeCheckoutAgainstTestMode(t *testing.T) {
	secretKey := os.Getenv("STRIPE_TEST_SECRET_KEY")
	if !strings.HasPrefix(secretKey, "sk_test_") {
		t.Skip("set STRIPE_TEST_SECRET_KEY to an sk_test_ key to run against Stripe test mode")
	}
	checkout := payments.NewStripeCheckout(secretKey, "http://localhost:3000", time.Now)
	paymentID, _ := uuid.NewV7()

	session, err := checkout.CreateSession(context.Background(), payments.CheckoutRequest{
		PaymentID:   paymentID,
		Description: "Nebula Exchange entry fee",
		AmountCents: 500,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if !strings.HasPrefix(session.ProviderSessionID, "cs_test_") || !strings.HasPrefix(session.URL, "https://checkout.stripe.com/") {
		t.Fatalf("unexpected session %+v", session)
	}
	if err := checkout.CancelSession(context.Background(), session.ProviderSessionID); err != nil {
		t.Fatalf("cancel session: %v", err)
	}
	if err := checkout.CancelSession(context.Background(), session.ProviderSessionID); err != nil && !errors.Is(err, payments.ErrCheckoutAlreadyPaid) {
		t.Fatalf("cancelling twice must be harmless: %v", err)
	}
}
