package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/wallet/repository"
)

type LinkedWalletLookup interface {
	LinkedWalletAddress(ctx context.Context, userID uuid.UUID) (string, bool, error)
}

type linkedWalletLookup struct {
	wallets repository.WalletRepository
}

func NewLinkedWalletLookup(wallets repository.WalletRepository) LinkedWalletLookup {
	return &linkedWalletLookup{wallets: wallets}
}

func (lookup *linkedWalletLookup) LinkedWalletAddress(ctx context.Context, userID uuid.UUID) (string, bool, error) {
	wallet, err := lookup.wallets.FindByUser(ctx, userID)
	if errors.Is(err, repository.ErrWalletNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return wallet.Address, true, nil
}
