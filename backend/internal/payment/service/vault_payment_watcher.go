package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment/repository"
)

const (
	chapterPaymentVaultCursorName = "chapter_payment_vault"
	maximumBlocksPerSync          = 500
)

type VaultPaymentWatcherDependencies struct {
	Chain                 VaultPaymentReader
	Cursors               repository.ChainSyncCursorRepository
	Payments              repository.PaymentRepository
	Unlocker              ChapterUnlocker
	Transactions          TransactionRunner
	RequiredConfirmations int
	PollInterval          time.Duration
	Logger                *slog.Logger
	Now                   Clock
}

type VaultPaymentWatcher struct {
	dependencies VaultPaymentWatcherDependencies
	settlement   walletPaymentSettlement
}

func NewVaultPaymentWatcher(dependencies VaultPaymentWatcherDependencies) *VaultPaymentWatcher {
	return &VaultPaymentWatcher{
		dependencies: dependencies,
		settlement: walletPaymentSettlement{
			payments:              dependencies.Payments,
			unlocker:              dependencies.Unlocker,
			transactions:          dependencies.Transactions,
			logger:                dependencies.Logger,
			now:                   dependencies.Now,
			requiredConfirmations: dependencies.RequiredConfirmations,
		},
	}
}

func (watcher *VaultPaymentWatcher) Run(ctx context.Context) {
	watcher.dependencies.Logger.Info("vault payment watcher started", slog.Duration("poll_interval", watcher.dependencies.PollInterval))
	pollTicker := time.NewTicker(watcher.dependencies.PollInterval)
	defer pollTicker.Stop()
	for {
		if _, err := watcher.SyncOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
			watcher.dependencies.Logger.Warn("vault payment sync failed", slog.Any("error", err))
		}
		select {
		case <-ctx.Done():
			watcher.dependencies.Logger.Info("vault payment watcher stopped")
			return
		case <-pollTicker.C:
		}
	}
}

func (watcher *VaultPaymentWatcher) SyncOnce(ctx context.Context) (int, error) {
	chain := watcher.dependencies.Chain
	headBlock, err := chain.HeadBlockNumber(ctx)
	if err != nil {
		return 0, err
	}
	pendingBlocks := uint64(watcher.dependencies.RequiredConfirmations - 1)
	if headBlock < pendingBlocks {
		return 0, nil
	}
	confirmedHead := headBlock - pendingBlocks

	cursor, err := watcher.dependencies.Cursors.Find(ctx, chapterPaymentVaultCursorName)
	isFirstSync := errors.Is(err, repository.ErrChainSyncCursorNotFound)
	if err != nil && !isFirstSync {
		return 0, err
	}
	if isFirstSync {
		return 0, watcher.saveCursor(ctx, confirmedHead)
	}
	isStillValid, err := watcher.isCursorOnChain(ctx, cursor, confirmedHead)
	if err != nil {
		return 0, err
	}
	if !isStillValid {
		watcher.dependencies.Logger.Warn("chain was reset or reorganized, watching from the current block", slog.Int64("previous_block", cursor.BlockNumber), slog.Uint64("current_block", confirmedHead))
		return 0, watcher.saveCursor(ctx, confirmedHead)
	}

	fromBlock := uint64(cursor.BlockNumber) + 1
	toBlock := min(confirmedHead, fromBlock+maximumBlocksPerSync-1)
	if fromBlock > toBlock {
		return 0, nil
	}
	observations, err := chain.PaymentsInBlockRange(ctx, fromBlock, toBlock)
	if err != nil {
		return 0, err
	}
	for _, observation := range observations {
		if err := watcher.settlement.settle(ctx, observation, headBlock); err != nil {
			return 0, err
		}
	}
	return len(observations), watcher.saveCursor(ctx, toBlock)
}

func (watcher *VaultPaymentWatcher) isCursorOnChain(ctx context.Context, cursor models.ChainSyncCursor, confirmedHead uint64) (bool, error) {
	if cursor.ChainID != watcher.dependencies.Chain.ChainID() || cursor.BlockNumber < 0 || uint64(cursor.BlockNumber) > confirmedHead {
		return false, nil
	}
	blockHash, err := watcher.dependencies.Chain.BlockHash(ctx, uint64(cursor.BlockNumber))
	if err != nil {
		return false, err
	}
	return blockHash == cursor.BlockHash, nil
}

func (watcher *VaultPaymentWatcher) saveCursor(ctx context.Context, blockNumber uint64) error {
	blockHash, err := watcher.dependencies.Chain.BlockHash(ctx, blockNumber)
	if err != nil {
		return err
	}
	return watcher.dependencies.Cursors.Save(ctx, &models.ChainSyncCursor{
		Name:        chapterPaymentVaultCursorName,
		ChainID:     watcher.dependencies.Chain.ChainID(),
		BlockNumber: int64(blockNumber),
		BlockHash:   blockHash,
		UpdatedAt:   watcher.dependencies.Now().UTC(),
	})
}
