package inventory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/ledger/ledgerstore"
	"nebula-exchange/backend/internal/platform/pagination"
)

type Holding struct {
	ItemID    int   `json:"item_id"`
	Available int64 `json:"available"`
	Held      int64 `json:"held"`
}

type Cursor struct {
	AfterItemID int `json:"after_item_id"`
}

type Reader struct {
	pool *pgxpool.Pool
}

func NewReader(pool *pgxpool.Pool) *Reader {
	return &Reader{pool: pool}
}

func (reader *Reader) List(ctx context.Context, userID uuid.UUID, pageRequest pagination.Request, cursor Cursor) (pagination.Page[Holding], error) {
	holdingRows, err := ledgerstore.New(reader.pool).ListPlayerInventory(ctx, ledgerstore.ListPlayerInventoryParams{
		UserID:      &userID,
		AfterItemID: int32(cursor.AfterItemID),
		RowLimit:    int32(pageRequest.FetchLimit()),
	})
	if err != nil {
		return pagination.Page[Holding]{}, fmt.Errorf("list inventory: %w", err)
	}
	holdings := make([]Holding, 0, len(holdingRows))
	for _, holdingRow := range holdingRows {
		holdings = append(holdings, Holding{ItemID: int(holdingRow.ItemID), Available: holdingRow.Available, Held: holdingRow.Held})
	}
	return pagination.BuildPage(holdings, pageRequest, func(holding Holding) Cursor {
		return Cursor{AfterItemID: holding.ItemID}
	})
}
