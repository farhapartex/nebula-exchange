package payments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"

	"nebula-exchange/backend/internal/payments/paymentsstore"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/httpserver/response"
	"nebula-exchange/backend/internal/platform/money"
)

const (
	maximumWebhookBodyBytes = 1 << 20
	stripeEventIDPrefix     = "stripe:"
)

type StripeWebhookHandler struct {
	pool          *pgxpool.Pool
	settler       *Settler
	webhookSecret string
	logger        *slog.Logger
	now           func() time.Time
}

func NewStripeWebhookHandler(pool *pgxpool.Pool, settler *Settler, webhookSecret string, logger *slog.Logger, now func() time.Time) *StripeWebhookHandler {
	return &StripeWebhookHandler{pool: pool, settler: settler, webhookSecret: webhookSecret, logger: logger, now: now}
}

func (handler *StripeWebhookHandler) RegisterRoutes(router gin.IRouter) {
	router.POST("/webhooks/stripe", handler.receive)
}

func (handler *StripeWebhookHandler) receive(context *gin.Context) {
	payload, err := io.ReadAll(io.LimitReader(context.Request.Body, maximumWebhookBodyBytes))
	if err != nil {
		response.WriteError(context, apierror.BadRequest("Request body could not be read"))
		return
	}
	event, err := webhook.ConstructEventWithOptions(payload, context.GetHeader("Stripe-Signature"), handler.webhookSecret,
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true})
	if err != nil {
		handler.logger.Warn("rejected stripe webhook", slog.String("error", err.Error()))
		response.WriteError(context, apierror.BadRequest("Webhook signature is invalid"))
		return
	}

	if err := handler.handleEvent(context.Request.Context(), event); err != nil {
		handler.logger.Error("stripe webhook failed", slog.String("event_id", event.ID), slog.String("type", string(event.Type)), slog.String("error", err.Error()))
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, map[string]bool{"received": true})
}

func (handler *StripeWebhookHandler) handleEvent(ctx context.Context, event stripe.Event) error {
	switch event.Type {
	case "checkout.session.completed", "checkout.session.async_payment_succeeded":
		session, paymentID, isOurs := handler.decodeSession(event)
		if !isOurs || session.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
			return nil
		}
		outcome, err := handler.settler.Settle(ctx, SettlementRequest{
			PaymentID: paymentID,
			Credited:  money.Micro(session.AmountTotal * microPerCent),
			Event: ExternalEvent{
				ID:       stripeEventIDPrefix + event.ID,
				Provider: "stripe",
				Type:     string(event.Type),
				Payload:  map[string]any{"session_id": session.ID, "amount_total": session.AmountTotal, "currency": session.Currency},
			},
		})
		if err != nil {
			return err
		}
		handler.logger.Info("stripe checkout settled", slog.String("payment_id", paymentID.String()), slog.Any("outcome", outcome))
		return nil
	case "checkout.session.expired", "checkout.session.async_payment_failed":
		_, paymentID, isOurs := handler.decodeSession(event)
		if !isOurs {
			return nil
		}
		return handler.expirePayment(ctx, event, paymentID)
	default:
		return nil
	}
}

func (handler *StripeWebhookHandler) decodeSession(event stripe.Event) (stripe.CheckoutSession, uuid.UUID, bool) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		handler.logger.Warn("stripe event without a checkout session", slog.String("event_id", event.ID))
		return session, uuid.Nil, false
	}
	paymentID, err := uuid.Parse(session.Metadata["payment_id"])
	if err != nil {
		handler.logger.Warn("stripe checkout session without our payment id", slog.String("event_id", event.ID), slog.String("session_id", session.ID))
		return session, uuid.Nil, false
	}
	return session, paymentID, true
}

func (handler *StripeWebhookHandler) expirePayment(ctx context.Context, event stripe.Event, paymentID uuid.UUID) error {
	return pgx.BeginFunc(ctx, handler.pool, func(tx pgx.Tx) error {
		queries := paymentsstore.New(tx)
		eventID := stripeEventIDPrefix + event.ID
		if _, err := queries.InsertExternalEvent(ctx, paymentsstore.InsertExternalEventParams{
			ID: eventID, Provider: "stripe", EventType: string(event.Type), Payload: []byte(`{}`),
		}); errors.Is(err, pgx.ErrNoRows) {
			return nil
		} else if err != nil {
			return fmt.Errorf("record stripe event: %w", err)
		}
		if _, err := queries.ExpirePendingPayment(ctx, paymentsstore.ExpirePendingPaymentParams{ID: paymentID, ExpiredAt: handler.now()}); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("expire payment: %w", err)
		}
		return queries.MarkExternalEventProcessed(ctx, eventID)
	})
}
