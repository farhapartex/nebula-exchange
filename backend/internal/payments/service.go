package payments

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/payments/paymentsstore"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/money"
	"nebula-exchange/backend/internal/platform/pagination"
	"nebula-exchange/backend/internal/purpose"
	"nebula-exchange/backend/internal/shop"
	"nebula-exchange/backend/internal/users"
)

type CreateRequest struct {
	Purpose   purpose.Kind
	Method    Method
	AmountNC  string
	SKU       string
	UpgradeID string
	Quantity  int
}

type Dependencies struct {
	Pool     *pgxpool.Pool
	Users    *users.Repository
	Catalog  *catalog.Service
	Checkout CardCheckout
	Now      func() time.Time
}

type Service struct {
	dependencies Dependencies
	queries      *paymentsstore.Queries
}

func NewService(dependencies Dependencies) *Service {
	return &Service{dependencies: dependencies, queries: paymentsstore.New(dependencies.Pool)}
}

type pricedPayment struct {
	amount      money.Micro
	sku         *string
	upgradeID   *string
	quantity    int
	description string
}

func (service *Service) Create(ctx context.Context, userID uuid.UUID, createRequest CreateRequest) (Payment, error) {
	payer, isFound, err := service.dependencies.Users.FindByID(ctx, service.dependencies.Pool, userID)
	if err != nil {
		return Payment{}, err
	}
	if !isFound {
		return Payment{}, apierror.Unauthorized("Log in to continue")
	}
	if err := ensureAccountMayPay(payer.Status, createRequest.Purpose); err != nil {
		return Payment{}, err
	}
	if createRequest.Method == MethodCrypto {
		return Payment{}, apierror.New(http.StatusUnprocessableEntity, apierror.CodeWalletRequired, "Link a wallet before paying with USDC")
	}

	priced, err := service.price(ctx, userID, createRequest)
	if err != nil {
		return Payment{}, err
	}

	paymentID, err := uuid.NewV7()
	if err != nil {
		return Payment{}, err
	}
	session, err := service.dependencies.Checkout.CreateSession(ctx, CheckoutRequest{
		PaymentID:   paymentID,
		Description: priced.description,
		AmountCents: int64(priced.amount) / microPerCent,
	})
	if errors.Is(err, ErrCardPaymentsUnavailable) {
		return Payment{}, apierror.New(http.StatusServiceUnavailable, apierror.CodeServiceUnavailable, "Card payments are not available right now")
	}
	if err != nil {
		return Payment{}, fmt.Errorf("create checkout session: %w", err)
	}

	createdAt := service.dependencies.Now()
	paymentRow, err := service.queries.InsertPayment(ctx, paymentsstore.InsertPaymentParams{
		ID:                paymentID,
		UserID:            userID,
		Purpose:           string(createRequest.Purpose),
		Method:            string(createRequest.Method),
		AmountMicro:       int64(priced.amount),
		Sku:               priced.sku,
		UpgradeID:         priced.upgradeID,
		Quantity:          int32(priced.quantity),
		ProviderSessionID: &session.ProviderSessionID,
		CheckoutUrl:       &session.URL,
		ExpiresAt:         createdAt.Add(lifetimeFor(createRequest.Method)),
		CreatedAt:         createdAt,
	})
	if err != nil {
		return Payment{}, fmt.Errorf("insert payment: %w", err)
	}
	return paymentFromRow(paymentRow), nil
}

func ensureAccountMayPay(status users.Status, purposeKind purpose.Kind) error {
	switch {
	case purposeKind == purpose.KindEntryFee && status == users.StatusPendingPayment:
		return nil
	case purposeKind == purpose.KindEntryFee && status == users.StatusActive:
		return apierror.Conflict("Your account is already active")
	case purposeKind != purpose.KindEntryFee && status == users.StatusActive:
		return nil
	default:
		return users.NotActiveError(status)
	}
}

func (service *Service) price(ctx context.Context, userID uuid.UUID, createRequest CreateRequest) (pricedPayment, error) {
	switch createRequest.Purpose {
	case purpose.KindEntryFee:
		return pricedPayment{amount: purpose.EntryFee, quantity: 1, description: "Nebula Exchange entry fee"}, nil
	case purpose.KindTopup:
		amount, err := money.ParseNC(createRequest.AmountNC)
		if err != nil || !allowedCardTopupAmounts[amount] {
			return pricedPayment{}, apierror.ValidationFailed(map[string]string{"amount_nc": "must be 5, 10, 25, 50 or 100"})
		}
		if err := service.ensureWithinTopupLimits(ctx, userID, amount); err != nil {
			return pricedPayment{}, err
		}
		return pricedPayment{amount: amount, quantity: 1, description: fmt.Sprintf("%s NC top-up", createRequest.AmountNC)}, nil
	case purpose.KindShopPurchase:
		snapshot, err := service.dependencies.Catalog.Snapshot(ctx)
		if err != nil {
			return pricedPayment{}, err
		}
		shopItem, isListed := snapshot.ShopItemBySKU(createRequest.SKU)
		if !isListed {
			return pricedPayment{}, apierror.ValidationFailed(map[string]string{"sku": "is not sold in the shop"})
		}
		quantity := max(createRequest.Quantity, 1)
		if quantity > shop.MaximumQuantity {
			return pricedPayment{}, apierror.ValidationFailed(map[string]string{"quantity": fmt.Sprintf("must be from 1 to %d", shop.MaximumQuantity)})
		}
		sku := shopItem.SKU
		return pricedPayment{
			amount:      shopItem.Price * money.Micro(quantity),
			sku:         &sku,
			quantity:    quantity,
			description: fmt.Sprintf("%d × %s", quantity, shopItem.Name),
		}, nil
	case purpose.KindUpgradePurchase:
		snapshot, err := service.dependencies.Catalog.Snapshot(ctx)
		if err != nil {
			return pricedPayment{}, err
		}
		upgrade, isKnownUpgrade := snapshot.UpgradeByID(createRequest.UpgradeID)
		if !isKnownUpgrade || upgrade.BuyPrice == nil {
			return pricedPayment{}, apierror.ValidationFailed(map[string]string{"upgrade_id": "is not an upgrade you can buy"})
		}
		upgradeID := upgrade.ID
		toItem, _ := snapshot.ItemByID(upgrade.ToItemID)
		return pricedPayment{amount: *upgrade.BuyPrice, upgradeID: &upgradeID, quantity: 1, description: "Upgrade to " + toItem.Name}, nil
	default:
		return pricedPayment{}, apierror.ValidationFailed(map[string]string{"purpose": "must be ENTRY_FEE, TOPUP, SHOP_PURCHASE or UPGRADE_PURCHASE"})
	}
}

func (service *Service) ensureWithinTopupLimits(ctx context.Context, userID uuid.UUID, amount money.Micro) error {
	currentTime := service.dependencies.Now()
	for _, window := range []struct {
		name     string
		duration time.Duration
		limit    money.Micro
	}{
		{"daily", 24 * time.Hour, dailyCardTopupLimit},
		{"30_days", 30 * 24 * time.Hour, monthlyCardTopupLimit},
	} {
		alreadyToppedUp, err := service.queries.SumRecentCardTopups(ctx, paymentsstore.SumRecentCardTopupsParams{
			UserID: userID,
			Since:  currentTime.Add(-window.duration),
			Now:    currentTime,
		})
		if err != nil {
			return fmt.Errorf("sum recent top-ups: %w", err)
		}
		if money.Micro(alreadyToppedUp)+amount > window.limit {
			return apierror.New(http.StatusUnprocessableEntity, apierror.CodeLimitExceeded, "This top-up would go over your card limit").
				WithDetails(map[string]any{"window": window.name, "remaining": max(window.limit-money.Micro(alreadyToppedUp), 0)})
		}
	}
	return nil
}

func (service *Service) Get(ctx context.Context, userID, paymentID uuid.UUID) (Payment, error) {
	paymentRow, err := service.queries.GetPaymentForUser(ctx, paymentsstore.GetPaymentForUserParams{ID: paymentID, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return Payment{}, apierror.NotFound("This payment does not exist")
	}
	if err != nil {
		return Payment{}, fmt.Errorf("load payment: %w", err)
	}
	return paymentFromRow(paymentRow), nil
}

type ListCursor struct {
	BeforeID uuid.UUID `json:"before_id"`
}

func (service *Service) List(ctx context.Context, userID uuid.UUID, pageRequest pagination.Request, cursor *ListCursor) (pagination.Page[Payment], error) {
	parameters := paymentsstore.ListPaymentsForUserParams{UserID: userID, RowLimit: int32(pageRequest.FetchLimit())}
	if cursor != nil {
		parameters.BeforeID = &cursor.BeforeID
	}
	paymentRows, err := service.queries.ListPaymentsForUser(ctx, parameters)
	if err != nil {
		return pagination.Page[Payment]{}, fmt.Errorf("list payments: %w", err)
	}
	paymentList := make([]Payment, 0, len(paymentRows))
	for _, paymentRow := range paymentRows {
		paymentList = append(paymentList, paymentFromRow(paymentRow))
	}
	return pagination.BuildPage(paymentList, pageRequest, func(payment Payment) ListCursor {
		return ListCursor{BeforeID: payment.ID}
	})
}

func (service *Service) Cancel(ctx context.Context, userID, paymentID uuid.UUID) (Payment, error) {
	pendingPayment, err := service.Get(ctx, userID, paymentID)
	if err != nil {
		return Payment{}, err
	}
	if pendingPayment.Status != StatusPending {
		return pendingPayment, nil
	}
	paymentRow, err := service.queries.GetPaymentForUser(ctx, paymentsstore.GetPaymentForUserParams{ID: paymentID, UserID: userID})
	if err != nil {
		return Payment{}, fmt.Errorf("load payment: %w", err)
	}
	if paymentRow.ProviderSessionID != nil {
		cancelErr := service.dependencies.Checkout.CancelSession(ctx, *paymentRow.ProviderSessionID)
		if errors.Is(cancelErr, ErrCheckoutAlreadyPaid) {
			return Payment{}, apierror.Conflict("This payment was already completed").WithDetails(map[string]string{"status": "PAID"})
		}
		if cancelErr != nil {
			return Payment{}, fmt.Errorf("cancel checkout session: %w", cancelErr)
		}
	}
	failedAt := service.dependencies.Now()
	failedRow, err := service.queries.FailPendingPayment(ctx, paymentsstore.FailPendingPaymentParams{ID: paymentID, UserID: userID, FailedAt: failedAt})
	if errors.Is(err, pgx.ErrNoRows) {
		return service.Get(ctx, userID, paymentID)
	}
	if err != nil {
		return Payment{}, fmt.Errorf("fail payment: %w", err)
	}
	return paymentFromRow(failedRow), nil
}
