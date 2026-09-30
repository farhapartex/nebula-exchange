package shop

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/ledger"
)

const MaximumQuantity = 100

var ErrInvalidQuantity = errors.New("purchase quantity must be from 1 to 100")

type PurchaseRequest struct {
	UserID    uuid.UUID
	ShopItem  catalog.ShopItem
	Quantity  int
	Reference ledger.Reference
}

func Purchase(ctx context.Context, tx pgx.Tx, request PurchaseRequest) (uuid.UUID, error) {
	if request.Quantity <= 0 || request.Quantity > MaximumQuantity {
		return uuid.Nil, ErrInvalidQuantity
	}
	totalPrice := int64(request.ShopItem.Price) * int64(request.Quantity)
	paymentLegs, err := ledger.SpendLegs(ctx, tx, request.UserID, totalPrice)
	if err != nil {
		return uuid.Nil, err
	}
	legs := append(paymentLegs, ledger.Credit(ledger.System(ledger.SystemTreasury, ledger.NC), totalPrice))
	for _, content := range request.ShopItem.Contents {
		grantedQuantity := int64(content.Quantity) * int64(request.Quantity)
		legs = append(legs,
			ledger.Debit(ledger.System(ledger.SystemMint, ledger.Item(content.ItemID)), grantedQuantity),
			ledger.Credit(ledger.PlayerItem(request.UserID, content.ItemID), grantedQuantity),
		)
	}
	return ledger.Post(ctx, tx, ledger.Journal{
		Type:      ledger.JournalShopPurchase,
		Reference: request.Reference,
		Metadata:  map[string]any{"sku": request.ShopItem.SKU, "quantity": request.Quantity},
		Legs:      legs,
	})
}
