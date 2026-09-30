package purpose

import (
	"context"

	"github.com/jackc/pgx/v5"

	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/upgrades"
	"nebula-exchange/backend/internal/users"
)

type UpgradePurchaseHandler struct {
	catalog *catalog.Service
	users   *users.Repository
}

func NewUpgradePurchaseHandler(catalogService *catalog.Service, userRepository *users.Repository) *UpgradePurchaseHandler {
	return &UpgradePurchaseHandler{catalog: catalogService, users: userRepository}
}

func (handler *UpgradePurchaseHandler) Apply(ctx context.Context, tx pgx.Tx, payment Payment) error {
	buyer, isFound, err := handler.users.FindByID(ctx, tx, payment.UserID)
	if err != nil {
		return err
	}
	if !isFound || buyer.Status != users.StatusActive {
		return &FailureError{Code: "ACCOUNT_NOT_ACTIVE"}
	}
	snapshot, err := handler.catalog.Snapshot(ctx)
	if err != nil {
		return err
	}
	upgrade, isKnownUpgrade := snapshot.UpgradeByID(payment.UpgradeID)
	if !isKnownUpgrade || upgrade.BuyPrice == nil {
		return &FailureError{Code: "UPGRADE_UNAVAILABLE"}
	}
	if _, err := upgrades.ApplyPurchase(ctx, tx, payment.UserID, upgrade, paymentReference(payment)); err != nil {
		if code, isExpected := FailureCodeOf(err); isExpected {
			return &FailureError{Code: code}
		}
		return err
	}
	return nil
}
