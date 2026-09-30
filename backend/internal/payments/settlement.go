package payments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/payments/paymentsstore"
	"nebula-exchange/backend/internal/platform/money"
	"nebula-exchange/backend/internal/purpose"
)

type ExternalEvent struct {
	ID       string
	Provider string
	Type     string
	Payload  map[string]any
}

type SettlementRequest struct {
	PaymentID uuid.UUID
	Credited  money.Micro
	Event     ExternalEvent
}

type SettlementOutcome struct {
	WasDuplicateEvent bool
	WasSettled        bool
	PurposeApplied    bool
	PurposeFailure    string
}

type Settler struct {
	pool    *pgxpool.Pool
	purpose *purpose.Runner
	logger  *slog.Logger
	now     func() time.Time
}

func NewSettler(pool *pgxpool.Pool, purposeRunner *purpose.Runner, logger *slog.Logger, now func() time.Time) *Settler {
	return &Settler{pool: pool, purpose: purposeRunner, logger: logger, now: now}
}

func (settler *Settler) Settle(ctx context.Context, request SettlementRequest) (SettlementOutcome, error) {
	var outcome SettlementOutcome
	err := pgx.BeginFunc(ctx, settler.pool, func(tx pgx.Tx) error {
		queries := paymentsstore.New(tx)
		encodedPayload, err := json.Marshal(request.Event.Payload)
		if err != nil {
			return fmt.Errorf("encode event payload: %w", err)
		}
		if _, err := queries.InsertExternalEvent(ctx, paymentsstore.InsertExternalEventParams{
			ID:        request.Event.ID,
			Provider:  request.Event.Provider,
			EventType: request.Event.Type,
			Payload:   encodedPayload,
		}); errors.Is(err, pgx.ErrNoRows) {
			outcome.WasDuplicateEvent = true
			return nil
		} else if err != nil {
			return fmt.Errorf("record external event: %w", err)
		}

		claimedPayment, err := queries.ClaimPendingPayment(ctx, paymentsstore.ClaimPendingPaymentParams{
			ID:            request.PaymentID,
			CreditedMicro: (*int64)(&request.Credited),
			SucceededAt:   ptrTo(settler.now()),
		})
		if errors.Is(err, pgx.ErrNoRows) {
			settler.logger.Warn("payment event for a payment that is not pending", slog.String("payment_id", request.PaymentID.String()), slog.String("event_id", request.Event.ID))
			return queries.MarkExternalEventProcessed(ctx, request.Event.ID)
		}
		if err != nil {
			return fmt.Errorf("claim payment: %w", err)
		}
		outcome.WasSettled = true

		if err := creditPayer(ctx, tx, claimedPayment, request.Credited); err != nil {
			return err
		}
		purposeFailure, err := settler.applyPurpose(ctx, tx, claimedPayment)
		if err != nil {
			return err
		}
		outcome.PurposeApplied = purposeFailure == ""
		outcome.PurposeFailure = purposeFailure

		purposeStatus := "APPLIED"
		var failureCode *string
		if purposeFailure != "" {
			purposeStatus = "FAILED"
			failureCode = &purposeFailure
		}
		if err := queries.RecordPurposeOutcome(ctx, paymentsstore.RecordPurposeOutcomeParams{
			ID:                 claimedPayment.ID,
			PurposeStatus:      purposeStatus,
			PurposeFailureCode: failureCode,
		}); err != nil {
			return fmt.Errorf("record purpose outcome: %w", err)
		}
		return queries.MarkExternalEventProcessed(ctx, request.Event.ID)
	})
	return outcome, err
}

func creditPayer(ctx context.Context, tx pgx.Tx, payment paymentsstore.Payment, credited money.Micro) error {
	journalType, clearingAccount, bucket := ledger.JournalTopupCard, ledger.SystemStripeClearing, ledger.BucketCard
	if Method(payment.Method) == MethodCrypto {
		journalType, clearingAccount, bucket = ledger.JournalTopupCrypto, ledger.SystemCryptoClearing, ledger.BucketCrypto
	}
	_, err := ledger.Post(ctx, tx, ledger.Journal{
		Type:      journalType,
		Reference: ledger.Reference{Type: "payment", ID: payment.ID.String()},
		Legs: []ledger.Leg{
			ledger.Debit(ledger.System(clearingAccount, ledger.NC), int64(credited)),
			ledger.Credit(ledger.PlayerNC(payment.UserID, bucket), int64(credited)),
		},
	})
	if err != nil {
		return fmt.Errorf("credit payer: %w", err)
	}
	return nil
}

func (settler *Settler) applyPurpose(ctx context.Context, tx pgx.Tx, payment paymentsstore.Payment) (string, error) {
	purposePayment := purpose.Payment{ID: payment.ID, UserID: payment.UserID, Kind: purpose.Kind(payment.Purpose), Quantity: int(payment.Quantity)}
	if payment.Sku != nil {
		purposePayment.SKU = *payment.Sku
	}
	if payment.UpgradeID != nil {
		purposePayment.UpgradeID = *payment.UpgradeID
	}

	savepoint, err := tx.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("open purpose savepoint: %w", err)
	}
	purposeErr := settler.purpose.Run(ctx, savepoint, purposePayment)
	if purposeErr == nil {
		return "", savepoint.Commit(ctx)
	}
	if rollbackErr := savepoint.Rollback(ctx); rollbackErr != nil {
		return "", fmt.Errorf("roll back purpose: %w", rollbackErr)
	}
	failureCode, isExpectedFailure := purpose.FailureCodeOf(purposeErr)
	if !isExpectedFailure {
		return "", fmt.Errorf("apply purpose: %w", purposeErr)
	}
	settler.logger.Info("payment purpose failed, NC kept in balance",
		slog.String("payment_id", payment.ID.String()), slog.String("failure", failureCode))
	return failureCode, nil
}

func ptrTo[Value any](value Value) *Value {
	return &value
}
