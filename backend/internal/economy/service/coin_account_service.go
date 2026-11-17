package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/economy/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/economy/repository"
)

const WelcomeGiftCoins int64 = 100

var ErrCoinAmountNotPositive = errors.New("coin amount must be positive")

type CoinAccountService interface {
	Balance(ctx context.Context, userID uuid.UUID) (int64, error)
	GrantWelcomeGift(ctx context.Context, userID uuid.UUID) error
	CreditCoinPack(ctx context.Context, userID uuid.UUID, paymentID uuid.UUID, coins int64) error
	ClawBackCoinPack(ctx context.Context, userID uuid.UUID, paymentID uuid.UUID, coins int64) error
}

type CoinAccountDependencies struct {
	Balances      repository.CoinBalanceRepository
	Ledger        repository.CoinLedgerRepository
	ClawbackFlags repository.CoinClawbackFlagRepository
	Transactions  TransactionRunner
	Logger        *slog.Logger
	Now           Clock
}

type coinAccountService struct {
	dependencies CoinAccountDependencies
}

func NewCoinAccountService(dependencies CoinAccountDependencies) CoinAccountService {
	return &coinAccountService{dependencies: dependencies}
}

func (accounts *coinAccountService) Balance(ctx context.Context, userID uuid.UUID) (int64, error) {
	return accounts.dependencies.Balances.CoinsOf(ctx, userID)
}

func (accounts *coinAccountService) GrantWelcomeGift(ctx context.Context, userID uuid.UUID) error {
	return accounts.dependencies.Transactions.WithinTransaction(ctx, func(ctx context.Context) error {
		balance, changedAt, err := accounts.lockBalance(ctx, userID)
		if err != nil {
			return err
		}
		hasWelcomeGift, err := accounts.dependencies.Ledger.HasWelcomeGift(ctx, userID)
		if err != nil || hasWelcomeGift {
			return err
		}
		return accounts.appendEntry(ctx, balance, WelcomeGiftCoins, models.CoinEntryReasonWelcomeGift, nil, changedAt)
	})
}

func (accounts *coinAccountService) CreditCoinPack(ctx context.Context, userID uuid.UUID, paymentID uuid.UUID, coins int64) error {
	if coins <= 0 {
		return ErrCoinAmountNotPositive
	}
	return accounts.dependencies.Transactions.WithinTransaction(ctx, func(ctx context.Context) error {
		balance, changedAt, err := accounts.lockBalance(ctx, userID)
		if err != nil {
			return err
		}
		isCredited, err := accounts.dependencies.Ledger.HasPaymentEntry(ctx, paymentID, models.CoinEntryReasonCoinPack)
		if err != nil || isCredited {
			return err
		}
		accounts.dependencies.Logger.InfoContext(ctx, "coin pack credited", slog.String("payment_id", paymentID.String()), slog.Int64("coins", coins))
		return accounts.appendEntry(ctx, balance, coins, models.CoinEntryReasonCoinPack, &paymentID, changedAt)
	})
}

func (accounts *coinAccountService) ClawBackCoinPack(ctx context.Context, userID uuid.UUID, paymentID uuid.UUID, coins int64) error {
	if coins <= 0 {
		return ErrCoinAmountNotPositive
	}
	return accounts.dependencies.Transactions.WithinTransaction(ctx, func(ctx context.Context) error {
		balance, changedAt, err := accounts.lockBalance(ctx, userID)
		if err != nil {
			return err
		}
		isAlreadyHandled, err := accounts.isClawbackHandled(ctx, paymentID)
		if err != nil || isAlreadyHandled {
			return err
		}
		clawedBackCoins := min(coins, balance.Coins)
		if clawedBackCoins > 0 {
			if err := accounts.appendEntry(ctx, balance, -clawedBackCoins, models.CoinEntryReasonCoinPackClawback, &paymentID, changedAt); err != nil {
				return err
			}
		}
		missingCoins := coins - clawedBackCoins
		accounts.dependencies.Logger.InfoContext(ctx, "coin pack clawed back",
			slog.String("payment_id", paymentID.String()),
			slog.Int64("clawed_back_coins", clawedBackCoins),
			slog.Int64("missing_coins", missingCoins),
		)
		if missingCoins == 0 {
			return nil
		}
		return accounts.flagForReview(ctx, userID, paymentID, coins, clawedBackCoins, changedAt)
	})
}

func (accounts *coinAccountService) isClawbackHandled(ctx context.Context, paymentID uuid.UUID) (bool, error) {
	isClawedBack, err := accounts.dependencies.Ledger.HasPaymentEntry(ctx, paymentID, models.CoinEntryReasonCoinPackClawback)
	if err != nil || isClawedBack {
		return isClawedBack, err
	}
	return accounts.dependencies.ClawbackFlags.ExistsForPayment(ctx, paymentID)
}

func (accounts *coinAccountService) flagForReview(ctx context.Context, userID uuid.UUID, paymentID uuid.UUID, packCoins int64, clawedBackCoins int64, flaggedAt time.Time) error {
	flagID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	accounts.dependencies.Logger.WarnContext(ctx, "coin pack reversed after its coins were spent", slog.String("payment_id", paymentID.String()), slog.String("user_id", userID.String()))
	return accounts.dependencies.ClawbackFlags.Create(ctx, &models.CoinClawbackFlag{
		ID:              flagID,
		UserID:          userID,
		PaymentID:       paymentID,
		PackCoins:       packCoins,
		ClawedBackCoins: clawedBackCoins,
		MissingCoins:    packCoins - clawedBackCoins,
		CreatedAt:       flaggedAt,
	})
}

func (accounts *coinAccountService) lockBalance(ctx context.Context, userID uuid.UUID) (models.CoinBalance, time.Time, error) {
	changedAt := accounts.dependencies.Now().UTC()
	balance, err := accounts.dependencies.Balances.LockOrCreate(ctx, userID, changedAt)
	return balance, changedAt, err
}

func (accounts *coinAccountService) appendEntry(ctx context.Context, balance models.CoinBalance, amount int64, reason models.CoinEntryReason, paymentID *uuid.UUID, changedAt time.Time) error {
	entryID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	balanceAfter := balance.Coins + amount
	if err := accounts.dependencies.Ledger.Append(ctx, &models.CoinLedgerEntry{
		ID:           entryID,
		UserID:       balance.UserID,
		Amount:       amount,
		BalanceAfter: balanceAfter,
		Reason:       reason,
		PaymentID:    paymentID,
		CreatedAt:    changedAt,
	}); err != nil {
		return err
	}
	return accounts.dependencies.Balances.SetCoins(ctx, balance.UserID, balanceAfter, changedAt)
}
