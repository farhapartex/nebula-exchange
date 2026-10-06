package service

import (
	"context"
	"encoding/hex"
	"errors"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
)

const (
	walletCheckoutLifetime        = 30 * time.Minute
	reusableAuthorizationLifetime = 10 * time.Minute
	walletLatePaymentGrace        = 2 * time.Minute
)

var (
	ErrWalletPaymentsUnavailable = apierror.ServiceUnavailable("Wallet payments are not available right now. Try again later.")
	ErrWalletRequired            = apierror.New(403, apierror.CodeWalletRequired, "Link a wallet before paying with it")
	ErrInvalidTransactionHash    = apierror.ValidationFailed(map[string]string{"transaction_hash": "must be a transaction hash"})
	ErrNotAWalletCheckout        = apierror.Conflict("This checkout is not a wallet payment")
	transactionHashPattern       = regexp.MustCompile(`^0x[0-9a-fA-F]{64}$`)
)

type StartedWalletCheckout struct {
	ID               uuid.UUID
	PaymentReference string
	USDCents         int64
	Deadline         time.Time
	Signature        string
	VaultAddress     string
	ChainID          int64
	IsReused         bool
}

func (checkout *checkoutService) StartWallet(ctx context.Context, userID uuid.UUID, planID string, chapterCount int) (StartedWalletCheckout, error) {
	walletPayments := checkout.dependencies.WalletPayments
	if !walletPayments.isEnabled() {
		return StartedWalletCheckout{}, ErrWalletPaymentsUnavailable
	}
	payerAddress, isLinked, err := walletPayments.LinkedWallets.LinkedWalletAddress(ctx, userID)
	if err != nil {
		return StartedWalletCheckout{}, err
	}
	if !isLinked {
		return StartedWalletCheckout{}, ErrWalletRequired
	}
	plan, option, err := checkout.quote(ctx, userID, planID, chapterCount)
	if err != nil {
		return StartedWalletCheckout{}, err
	}
	reusablePayment, err := checkout.findReusableWalletCheckout(ctx, userID, plan.ID, option, payerAddress)
	if err != nil {
		return StartedWalletCheckout{}, err
	}
	if reusablePayment != nil {
		startedCheckout := checkout.startedWalletCheckoutFrom(*reusablePayment)
		startedCheckout.IsReused = true
		return startedCheckout, nil
	}
	if err := checkout.expireOpenCheckouts(ctx, userID); err != nil {
		return StartedWalletCheckout{}, err
	}

	payment, err := checkout.newPayment(userID, plan, option)
	if err != nil {
		return StartedWalletCheckout{}, err
	}
	paymentReference := "0x" + hex.EncodeToString(crypto.Keccak256([]byte(payment.ID.String())))
	deadline := payment.CreatedAt.Add(walletCheckoutLifetime).Truncate(time.Second)
	signature, err := walletPayments.Authorizer.SignPaymentAuthorization(paymentReference, payerAddress, payment.TotalCents, deadline)
	if err != nil {
		return StartedWalletCheckout{}, err
	}
	payment.PaymentMethod = models.PaymentMethodWallet
	payment.CheckoutExpiresAt = deadline
	payment.PaymentReference = &paymentReference
	payment.PayerAddress = &payerAddress
	payment.AuthorizationSignature = &signature
	if err := checkout.dependencies.Payments.CreateWithChapters(ctx, &payment); err != nil {
		if errors.Is(err, repository.ErrOpenCheckoutExists) {
			return StartedWalletCheckout{}, ErrCheckoutAlreadyStarting
		}
		return StartedWalletCheckout{}, err
	}
	return checkout.startedWalletCheckoutFrom(payment), nil
}

func (checkout *checkoutService) findReusableWalletCheckout(ctx context.Context, userID uuid.UUID, planID string, option PlanOption, payerAddress string) (*models.Payment, error) {
	openPayments, err := checkout.dependencies.Payments.ListOpenForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	offeredChapterIDs := make([]string, 0, len(option.Chapters))
	for _, chapter := range option.Chapters {
		offeredChapterIDs = append(offeredChapterIDs, chapter.ID)
	}
	minimumDeadline := checkout.dependencies.Now().Add(reusableAuthorizationLifetime)
	for _, openPayment := range openPayments {
		isSameWalletCheckout := openPayment.PaymentMethod == models.PaymentMethodWallet &&
			openPayment.PlanID == planID &&
			openPayment.TotalCents == option.TotalCents &&
			openPayment.PayerAddress != nil && *openPayment.PayerAddress == payerAddress &&
			openPayment.CheckoutExpiresAt.After(minimumDeadline) &&
			slices.Equal(openPayment.ChapterIDs(), offeredChapterIDs)
		if isSameWalletCheckout {
			return &openPayment, nil
		}
	}
	return nil, nil
}

func (checkout *checkoutService) startedWalletCheckoutFrom(payment models.Payment) StartedWalletCheckout {
	walletPayments := checkout.dependencies.WalletPayments
	return StartedWalletCheckout{
		ID:               payment.ID,
		PaymentReference: *payment.PaymentReference,
		USDCents:         payment.TotalCents,
		Deadline:         payment.CheckoutExpiresAt,
		Signature:        *payment.AuthorizationSignature,
		VaultAddress:     walletPayments.Authorizer.VaultAddress(),
		ChainID:          walletPayments.Authorizer.ChainID(),
	}
}

func (checkout *checkoutService) ReportWalletTransaction(ctx context.Context, userID uuid.UUID, paymentID uuid.UUID, transactionHash string) (CheckoutState, error) {
	if !transactionHashPattern.MatchString(transactionHash) {
		return CheckoutState{}, ErrInvalidTransactionHash
	}
	if !checkout.dependencies.WalletPayments.isEnabled() {
		return CheckoutState{}, ErrWalletPaymentsUnavailable
	}
	payment, err := checkout.findOwnPayment(ctx, userID, paymentID)
	if err != nil {
		return CheckoutState{}, err
	}
	if payment.PaymentMethod != models.PaymentMethodWallet {
		return CheckoutState{}, ErrNotAWalletCheckout
	}
	normalizedHash := strings.ToLower(transactionHash)
	if payment.Status == models.PaymentStatusOpen {
		if err := checkout.dependencies.Payments.Update(ctx, &payment, map[string]any{"reported_transaction_hash": normalizedHash}); err != nil {
			return CheckoutState{}, err
		}
	}
	if err := checkout.observeTransaction(ctx, payment, normalizedHash); err != nil {
		return CheckoutState{}, err
	}
	return checkout.walletState(ctx, userID, paymentID)
}

func (checkout *checkoutService) observeTransaction(ctx context.Context, payment models.Payment, transactionHash string) error {
	chain := checkout.dependencies.WalletPayments.Chain
	observations, err := chain.PaymentsInTransaction(ctx, transactionHash)
	if err != nil {
		checkout.dependencies.Logger.WarnContext(ctx, "wallet transaction could not be read", slog.String("payment_id", payment.ID.String()), slog.Any("error", err))
		return nil
	}
	if len(observations) == 0 {
		return nil
	}
	headBlock, err := chain.HeadBlockNumber(ctx)
	if err != nil {
		checkout.dependencies.Logger.WarnContext(ctx, "chain head could not be read", slog.Any("error", err))
		return nil
	}
	for _, observation := range observations {
		if payment.PaymentReference != nil && observation.PaymentReference == *payment.PaymentReference {
			if err := checkout.walletSettlement.settle(ctx, observation, headBlock); err != nil {
				return err
			}
		}
	}
	return nil
}

func (checkout *checkoutService) walletState(ctx context.Context, userID uuid.UUID, paymentID uuid.UUID) (CheckoutState, error) {
	payment, err := checkout.findOwnPayment(ctx, userID, paymentID)
	if err != nil {
		return CheckoutState{}, err
	}
	requiredConfirmations := checkout.walletSettlement.requiredConfirmations
	if payment.Status == models.PaymentStatusOpen && checkout.dependencies.WalletPayments.isEnabled() {
		if payment, err = checkout.refreshOpenWalletPayment(ctx, userID, payment); err != nil {
			return CheckoutState{}, err
		}
	}
	confirmations := 0
	switch {
	case payment.Status.WasPaid():
		confirmations = requiredConfirmations
	case payment.BlockNumber != nil && checkout.dependencies.WalletPayments.isEnabled():
		if headBlock, err := checkout.dependencies.WalletPayments.Chain.HeadBlockNumber(ctx); err == nil {
			confirmations = min(confirmationsOf(uint64(*payment.BlockNumber), headBlock), requiredConfirmations)
		}
	}
	return CheckoutState{
		ID:                    payment.ID,
		Status:                checkoutStatusOf(payment.Status),
		Confirmations:         &confirmations,
		RequiredConfirmations: &requiredConfirmations,
	}, nil
}

func (checkout *checkoutService) refreshOpenWalletPayment(ctx context.Context, userID uuid.UUID, payment models.Payment) (models.Payment, error) {
	if payment.ReportedTransactionHash != nil {
		if err := checkout.observeTransaction(ctx, payment, *payment.ReportedTransactionHash); err != nil {
			return models.Payment{}, err
		}
	}
	isPastDeadline := checkout.dependencies.Now().After(payment.CheckoutExpiresAt.Add(walletLatePaymentGrace))
	if isPastDeadline && payment.TransactionHash == nil {
		if err := checkout.settlement.markExpired(ctx, payment.ID); err != nil {
			return models.Payment{}, err
		}
	}
	return checkout.findOwnPayment(ctx, userID, payment.ID)
}

func (checkout *checkoutService) findOwnPayment(ctx context.Context, userID uuid.UUID, paymentID uuid.UUID) (models.Payment, error) {
	payment, err := checkout.dependencies.Payments.FindForUser(ctx, paymentID, userID)
	if errors.Is(err, repository.ErrPaymentNotFound) {
		return models.Payment{}, ErrCheckoutNotFound
	}
	return payment, err
}
