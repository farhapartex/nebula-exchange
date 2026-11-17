package service

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
)

const (
	coinCheckoutReturnParameter = "coin_checkout"
	defaultCoinCheckoutPath     = "/fight"
)

var coinCheckoutReturnPaths = []string{"/fight", "/fight/history", "/market", "/subscription"}

func (checkout *checkoutService) StartCoinPack(ctx context.Context, userID uuid.UUID, coinPackID string, returnPath string) (StartedCheckout, error) {
	if checkout.dependencies.Gateway == nil {
		return StartedCheckout{}, ErrPaymentsUnavailable
	}
	coinPack, err := checkout.dependencies.CoinPacks.FindActive(ctx, coinPackID)
	if errors.Is(err, repository.ErrCoinPackNotFound) {
		return StartedCheckout{}, apierror.ValidationFailed(map[string]string{"coin_pack_id": "is not available"})
	}
	if err != nil {
		return StartedCheckout{}, err
	}
	if err := checkout.expireOpenCheckouts(ctx, userID); err != nil {
		return StartedCheckout{}, err
	}
	payment, err := checkout.newCoinPackPayment(userID, coinPack)
	if err != nil {
		return StartedCheckout{}, err
	}
	safeReturnPath := coinCheckoutReturnPath(returnPath)
	return checkout.openStripeCheckout(ctx, &payment, stripeProduct{
		name:        fmt.Sprintf("%s: %d coins", checkoutProductNamePrefix, coinPack.Coins),
		description: coinPack.Name + ". Coins buy tools in the game and cannot be turned back into money.",
		successURL:  checkout.returnURL(safeReturnPath, coinCheckoutReturnParameter, payment.ID.String()),
		cancelURL:   checkout.returnURL(safeReturnPath, coinCheckoutReturnParameter, checkoutCancelledValue),
	})
}

func coinCheckoutReturnPath(requestedPath string) string {
	if slices.Contains(coinCheckoutReturnPaths, requestedPath) {
		return requestedPath
	}
	return defaultCoinCheckoutPath
}

func (checkout *checkoutService) newCoinPackPayment(userID uuid.UUID, coinPack models.CoinPack) (models.Payment, error) {
	paymentID, err := uuid.NewV7()
	if err != nil {
		return models.Payment{}, err
	}
	createdAt := checkout.dependencies.Now().UTC()
	coinPackID := coinPack.ID
	coins := coinPack.Coins
	return models.Payment{
		ID:                paymentID,
		UserID:            userID,
		CoinPackID:        &coinPackID,
		Coins:             &coins,
		Status:            models.PaymentStatusOpen,
		PaymentMethod:     models.PaymentMethodCard,
		Currency:          checkoutCurrency,
		SubtotalCents:     coinPack.PriceCents,
		DiscountPercent:   0,
		DiscountCents:     0,
		TotalCents:        coinPack.PriceCents,
		CheckoutExpiresAt: createdAt.Add(checkoutLifetime),
		CreatedAt:         createdAt,
		UpdatedAt:         createdAt,
	}, nil
}
