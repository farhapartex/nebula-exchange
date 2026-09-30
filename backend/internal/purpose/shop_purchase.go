package purpose

import (
	"context"

	"github.com/jackc/pgx/v5"

	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/shop"
	"nebula-exchange/backend/internal/users"
)

type ShopPurchaseHandler struct {
	catalog *catalog.Service
	users   *users.Repository
}

func NewShopPurchaseHandler(catalogService *catalog.Service, userRepository *users.Repository) *ShopPurchaseHandler {
	return &ShopPurchaseHandler{catalog: catalogService, users: userRepository}
}

func (handler *ShopPurchaseHandler) Apply(ctx context.Context, tx pgx.Tx, payment Payment) error {
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
	shopItem, isListed := snapshot.ShopItemBySKU(payment.SKU)
	if !isListed {
		return &FailureError{Code: "SKU_UNAVAILABLE"}
	}
	_, err = shop.Purchase(ctx, tx, shop.PurchaseRequest{
		UserID:    payment.UserID,
		ShopItem:  shopItem,
		Quantity:  1,
		Reference: paymentReference(payment),
	})
	return err
}
