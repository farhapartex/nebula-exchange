package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
)

var ErrInvalidWebhookSignature = apierror.BadRequest("The webhook signature is not valid")

type StripeWebhookService interface {
	Handle(ctx context.Context, payload []byte, signatureHeader string) error
}

type StripeWebhookDependencies struct {
	Verifier     PaymentEventVerifier
	Events       repository.StripeWebhookEventRepository
	Payments     repository.PaymentRepository
	Unlocker     ChapterUnlocker
	Transactions TransactionRunner
	Logger       *slog.Logger
	Now          Clock
}

type stripeWebhookService struct {
	dependencies StripeWebhookDependencies
	settlement   paymentSettlement
}

func NewStripeWebhookService(dependencies StripeWebhookDependencies) StripeWebhookService {
	return &stripeWebhookService{
		dependencies: dependencies,
		settlement: paymentSettlement{
			payments:     dependencies.Payments,
			unlocker:     dependencies.Unlocker,
			transactions: dependencies.Transactions,
			logger:       dependencies.Logger,
			now:          dependencies.Now,
		},
	}
}

func (webhook *stripeWebhookService) Handle(ctx context.Context, payload []byte, signatureHeader string) error {
	if webhook.dependencies.Verifier == nil {
		return ErrPaymentsUnavailable
	}
	event, err := webhook.dependencies.Verifier.VerifyEvent(payload, signatureHeader)
	if err != nil {
		webhook.dependencies.Logger.WarnContext(ctx, "stripe webhook rejected", slog.Any("error", err))
		return ErrInvalidWebhookSignature
	}
	return webhook.dependencies.Transactions.WithinTransaction(ctx, func(ctx context.Context) error {
		receivedAt := webhook.dependencies.Now().UTC()
		isNewEvent, err := webhook.dependencies.Events.RecordIfNew(ctx, &models.StripeWebhookEvent{
			ID:          event.ID,
			Type:        event.Type,
			Payload:     event.Payload,
			ReceivedAt:  receivedAt,
			ProcessedAt: &receivedAt,
		})
		if err != nil || !isNewEvent {
			return err
		}
		return webhook.apply(ctx, event)
	})
}

func (webhook *stripeWebhookService) apply(ctx context.Context, event PaymentEvent) error {
	switch event.Kind {
	case PaymentEventCheckoutCompleted:
		paymentID, isKnownPayment := webhook.paymentIDOf(ctx, event)
		if !isKnownPayment || !event.Checkout.IsPaid {
			return nil
		}
		return webhook.settlement.markPaid(ctx, paymentID, event.Checkout)
	case PaymentEventCheckoutExpired:
		paymentID, isKnownPayment := webhook.paymentIDOf(ctx, event)
		if !isKnownPayment {
			return nil
		}
		return webhook.settlement.markExpired(ctx, paymentID)
	case PaymentEventChargeRefunded:
		if !event.IsFullRefund || event.PaymentIntentID == "" {
			webhook.dependencies.Logger.InfoContext(ctx, "partial refund leaves chapters unlocked", slog.String("payment_intent_id", event.PaymentIntentID))
			return nil
		}
		return webhook.settlement.reverse(ctx, event.PaymentIntentID, models.PaymentStatusRefunded)
	case PaymentEventDisputeOpened:
		if event.PaymentIntentID == "" {
			return nil
		}
		return webhook.settlement.reverse(ctx, event.PaymentIntentID, models.PaymentStatusDisputed)
	}
	return nil
}

func (webhook *stripeWebhookService) paymentIDOf(ctx context.Context, event PaymentEvent) (uuid.UUID, bool) {
	paymentID, err := uuid.Parse(event.PaymentID)
	if err != nil {
		webhook.dependencies.Logger.WarnContext(ctx, "stripe checkout without a payment reference", slog.String("event_id", event.ID))
		return uuid.Nil, false
	}
	return paymentID, true
}
