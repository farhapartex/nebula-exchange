package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment/repository"
)

type walletPaymentSettlement struct {
	payments              repository.PaymentRepository
	unlocker              ChapterUnlocker
	transactions          TransactionRunner
	logger                *slog.Logger
	now                   Clock
	requiredConfirmations int
}

func confirmationsOf(blockNumber uint64, headBlock uint64) int {
	if headBlock < blockNumber {
		return 0
	}
	return int(headBlock - blockNumber + 1)
}

func (settlement walletPaymentSettlement) settle(ctx context.Context, observation ObservedVaultPayment, headBlock uint64) error {
	confirmations := confirmationsOf(observation.BlockNumber, headBlock)
	return settlement.transactions.WithinTransaction(ctx, func(ctx context.Context) error {
		payment, err := settlement.payments.LockByPaymentReference(ctx, observation.PaymentReference)
		if errors.Is(err, repository.ErrPaymentNotFound) {
			settlement.logger.WarnContext(ctx, "vault payment has no matching checkout", slog.String("payment_reference", observation.PaymentReference), slog.String("transaction_hash", observation.TransactionHash))
			return nil
		}
		if err != nil || payment.Status.WasPaid() {
			return err
		}
		if observation.USDCents < payment.TotalCents {
			settlement.logger.ErrorContext(ctx, "vault payment is below the checkout price", slog.String("payment_id", payment.ID.String()), slog.Int64("expected_cents", payment.TotalCents), slog.Int64("paid_cents", observation.USDCents))
			return nil
		}
		blockNumber := int64(observation.BlockNumber)
		if confirmations < settlement.requiredConfirmations {
			return settlement.payments.Update(ctx, &payment, map[string]any{"transaction_hash": observation.TransactionHash, "block_number": blockNumber})
		}
		paidAt := settlement.now().UTC()
		if err := settlement.payments.Update(ctx, &payment, map[string]any{
			"status":              models.PaymentStatusPaid,
			"paid_at":             paidAt,
			"transaction_hash":    observation.TransactionHash,
			"block_number":        blockNumber,
			"wallet_asset":        observation.Asset,
			"wallet_amount_units": observation.AmountUnits,
		}); err != nil {
			return err
		}
		settlement.logger.InfoContext(ctx, "wallet chapter payment paid", slog.String("payment_id", payment.ID.String()), slog.String("asset", string(observation.Asset)), slog.Int("chapters", len(payment.Chapters)))
		return settlement.unlocker.UnlockPurchasedChapters(ctx, payment.UserID, payment.ID, payment.ChapterIDs(), paidAt)
	})
}
