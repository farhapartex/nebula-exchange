package shop

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/money"
)

type CompletedPurchase struct {
	ID        uuid.UUID   `json:"id"`
	SKU       string      `json:"sku"`
	Quantity  int         `json:"quantity"`
	Total     money.Micro `json:"total"`
	JournalID uuid.UUID   `json:"journal_id"`
}

type Service struct {
	pool    *pgxpool.Pool
	catalog *catalog.Service
}

func NewService(pool *pgxpool.Pool, catalogService *catalog.Service) *Service {
	return &Service{pool: pool, catalog: catalogService}
}

func (service *Service) BuyWithBalance(ctx context.Context, userID uuid.UUID, sku string, quantity int) (CompletedPurchase, error) {
	snapshot, err := service.catalog.Snapshot(ctx)
	if err != nil {
		return CompletedPurchase{}, err
	}
	shopItem, isListed := snapshot.ShopItemBySKU(sku)
	if !isListed {
		return CompletedPurchase{}, apierror.NotFound("This item is not sold in the shop")
	}
	purchaseID, err := uuid.NewV7()
	if err != nil {
		return CompletedPurchase{}, err
	}

	var journalID uuid.UUID
	err = pgx.BeginFunc(ctx, service.pool, func(tx pgx.Tx) error {
		var purchaseErr error
		journalID, purchaseErr = Purchase(ctx, tx, PurchaseRequest{
			UserID:    userID,
			ShopItem:  shopItem,
			Quantity:  quantity,
			Reference: ledger.Reference{Type: "shop_purchase", ID: purchaseID.String()},
		})
		return purchaseErr
	})
	if err != nil {
		return CompletedPurchase{}, fmt.Errorf("buy %s: %w", sku, err)
	}
	return CompletedPurchase{
		ID:        purchaseID,
		SKU:       shopItem.SKU,
		Quantity:  quantity,
		Total:     shopItem.Price * money.Micro(quantity),
		JournalID: journalID,
	}, nil
}
