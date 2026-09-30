package purpose

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/users"
)

type Kind string

const (
	KindEntryFee        Kind = "ENTRY_FEE"
	KindTopup           Kind = "TOPUP"
	KindShopPurchase    Kind = "SHOP_PURCHASE"
	KindUpgradePurchase Kind = "UPGRADE_PURCHASE"
)

type Payment struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Kind      Kind
	SKU       string
	UpgradeID string
	Quantity  int
}

type Handler interface {
	Apply(ctx context.Context, tx pgx.Tx, payment Payment) error
}

type FailureError struct {
	Code string
}

func (failure *FailureError) Error() string {
	return "purpose could not be applied: " + failure.Code
}

func FailureCodeOf(err error) (string, bool) {
	var failure *FailureError
	if errors.As(err, &failure) {
		return failure.Code, true
	}
	var insufficientBalance *ledger.InsufficientBalanceError
	if errors.As(err, &insufficientBalance) {
		if insufficientBalance.Asset.IsNC() {
			return "INSUFFICIENT_FUNDS", true
		}
		return "INSUFFICIENT_ITEMS", true
	}
	return "", false
}

type Runner struct {
	handlers map[Kind]Handler
}

func NewRunner(handlers map[Kind]Handler) *Runner {
	return &Runner{handlers: handlers}
}

func (runner *Runner) Run(ctx context.Context, tx pgx.Tx, payment Payment) error {
	handler, isKnown := runner.handlers[payment.Kind]
	if !isKnown {
		return &FailureError{Code: "UNKNOWN_PURPOSE"}
	}
	return handler.Apply(ctx, tx, payment)
}

func paymentReference(payment Payment) ledger.Reference {
	return ledger.Reference{Type: "payment", ID: payment.ID.String()}
}

func NewDefaultRunner(userRepository *users.Repository, catalogService *catalog.Service, now func() time.Time) *Runner {
	return NewRunner(map[Kind]Handler{
		KindEntryFee:        NewEntryFeeHandler(userRepository, now),
		KindTopup:           TopupHandler{},
		KindShopPurchase:    NewShopPurchaseHandler(catalogService, userRepository),
		KindUpgradePurchase: NewUpgradePurchaseHandler(catalogService, userRepository),
	})
}
