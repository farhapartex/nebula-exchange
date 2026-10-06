package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment/repository"
)

type paymentSettlement struct {
	payments     repository.PaymentRepository
	unlocker     ChapterUnlocker
	transactions TransactionRunner
	logger       *slog.Logger
	now          Clock
}

func (settlement paymentSettlement) markPaid(ctx context.Context, paymentID uuid.UUID, checkout CheckoutSnapshot) error {
	return settlement.transactions.WithinTransaction(ctx, func(ctx context.Context) error {
		payment, err := settlement.payments.LockByID(ctx, paymentID)
		if errors.Is(err, repository.ErrPaymentNotFound) {
			settlement.logger.WarnContext(ctx, "paid checkout has no payment", slog.String("payment_id", paymentID.String()))
			return nil
		}
		if err != nil {
			return err
		}
		if payment.Status.WasPaid() {
			return nil
		}
		if !settlement.matchesPayment(ctx, payment, checkout) {
			return nil
		}
		paidAt := settlement.now().UTC()
		changes := map[string]any{"status": models.PaymentStatusPaid, "paid_at": paidAt}
		if checkout.PaymentIntentID != "" {
			changes["stripe_payment_intent_id"] = checkout.PaymentIntentID
		}
		if err := settlement.payments.Update(ctx, &payment, changes); err != nil {
			return err
		}
		settlement.logger.InfoContext(ctx, "chapter payment paid", slog.String("payment_id", payment.ID.String()), slog.Int("chapters", len(payment.Chapters)))
		return settlement.unlocker.UnlockPurchasedChapters(ctx, payment.UserID, payment.ID, payment.ChapterIDs(), paidAt)
	})
}

func (settlement paymentSettlement) matchesPayment(ctx context.Context, payment models.Payment, checkout CheckoutSnapshot) bool {
	isSameSession := payment.StripeCheckoutSessionID != nil && *payment.StripeCheckoutSessionID == checkout.SessionID
	isSameAmount := checkout.AmountTotalCents == payment.TotalCents && strings.EqualFold(checkout.Currency, payment.Currency)
	if isSameSession && isSameAmount {
		return true
	}
	settlement.logger.ErrorContext(ctx, "paid checkout does not match its payment",
		slog.String("payment_id", payment.ID.String()),
		slog.String("checkout_session_id", checkout.SessionID),
		slog.Int64("expected_cents", payment.TotalCents),
		slog.Int64("paid_cents", checkout.AmountTotalCents),
	)
	return false
}

func (settlement paymentSettlement) markExpired(ctx context.Context, paymentID uuid.UUID) error {
	return settlement.transactions.WithinTransaction(ctx, func(ctx context.Context) error {
		payment, err := settlement.payments.LockByID(ctx, paymentID)
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return nil
		}
		if err != nil || payment.Status != models.PaymentStatusOpen {
			return err
		}
		return settlement.payments.Update(ctx, &payment, map[string]any{"status": models.PaymentStatusExpired})
	})
}

func (settlement paymentSettlement) reverse(ctx context.Context, paymentIntentID string, reversedStatus models.PaymentStatus) error {
	return settlement.transactions.WithinTransaction(ctx, func(ctx context.Context) error {
		payment, err := settlement.payments.LockByPaymentIntent(ctx, paymentIntentID)
		if errors.Is(err, repository.ErrPaymentNotFound) {
			settlement.logger.WarnContext(ctx, "reversed charge has no payment", slog.String("payment_intent_id", paymentIntentID))
			return nil
		}
		if err != nil || payment.Status != models.PaymentStatusPaid {
			return err
		}
		if err := settlement.payments.Update(ctx, &payment, map[string]any{"status": reversedStatus, "refunded_at": settlement.now().UTC()}); err != nil {
			return err
		}
		if err := settlement.unlocker.LockPurchasedChapters(ctx, payment.ID); err != nil {
			return err
		}
		settlement.logger.InfoContext(ctx, "chapter payment reversed", slog.String("payment_id", payment.ID.String()), slog.String("status", string(reversedStatus)))
		return settlement.restoreChaptersPaidElsewhere(ctx, payment)
	})
}

func (settlement paymentSettlement) restoreChaptersPaidElsewhere(ctx context.Context, reversedPayment models.Payment) error {
	otherPayments, err := settlement.payments.ListPaidForUser(ctx, reversedPayment.UserID, reversedPayment.ID)
	if err != nil {
		return err
	}
	for _, otherPayment := range otherPayments {
		if err := settlement.unlocker.UnlockPurchasedChapters(ctx, otherPayment.UserID, otherPayment.ID, otherPayment.ChapterIDs(), *otherPayment.PaidAt); err != nil {
			return err
		}
	}
	return nil
}
