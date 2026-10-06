package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
)

type ObservedVaultPayment struct {
	PaymentReference string
	PayerAddress     string
	Asset            models.WalletPaymentAsset
	AmountUnits      string
	USDCents         int64
	TransactionHash  string
	BlockNumber      uint64
}

type VaultPaymentReader interface {
	ChainID() int64
	HeadBlockNumber(ctx context.Context) (uint64, error)
	BlockHash(ctx context.Context, blockNumber uint64) (string, error)
	PaymentsInBlockRange(ctx context.Context, fromBlock uint64, toBlock uint64) ([]ObservedVaultPayment, error)
	PaymentsInTransaction(ctx context.Context, transactionHash string) ([]ObservedVaultPayment, error)
}

type PaymentAuthorizer interface {
	VaultAddress() string
	ChainID() int64
	SignPaymentAuthorization(paymentReference string, payerAddress string, usdCents int64, deadline time.Time) (string, error)
}

type LinkedWalletLookup interface {
	LinkedWalletAddress(ctx context.Context, userID uuid.UUID) (string, bool, error)
}

type WalletPaymentDependencies struct {
	Authorizer            PaymentAuthorizer
	Chain                 VaultPaymentReader
	LinkedWallets         LinkedWalletLookup
	RequiredConfirmations int
}

func (walletPayments *WalletPaymentDependencies) isEnabled() bool {
	return walletPayments != nil && walletPayments.Authorizer != nil && walletPayments.Chain != nil && walletPayments.LinkedWallets != nil
}
