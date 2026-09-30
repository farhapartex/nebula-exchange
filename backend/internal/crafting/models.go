package crafting

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/crafting/craftingstore"
	"nebula-exchange/backend/internal/platform/money"
)

const MaximumBatchQuantity = 100

type Status string

const (
	StatusCrafting  Status = "CRAFTING"
	StatusDelivered Status = "DELIVERED"
)

type CraftJob struct {
	ID             uuid.UUID              `json:"id"`
	RecipeID       string                 `json:"recipe_id"`
	Quantity       int                    `json:"quantity"`
	OutputItemID   int                    `json:"output_item_id"`
	OutputQuantity int                    `json:"output_quantity"`
	Fee            money.Micro            `json:"fee"`
	Inputs         []catalog.ItemQuantity `json:"inputs"`
	Status         Status                 `json:"status"`
	StartedAt      time.Time              `json:"started_at"`
	EndsAt         time.Time              `json:"ends_at"`
	DeliveredAt    *time.Time             `json:"delivered_at"`
}

func craftJobFromRow(craftJobRow craftingstore.CraftJob) (CraftJob, error) {
	var inputs []catalog.ItemQuantity
	if err := json.Unmarshal(craftJobRow.Inputs, &inputs); err != nil {
		return CraftJob{}, err
	}
	return CraftJob{
		ID:             craftJobRow.ID,
		RecipeID:       craftJobRow.RecipeID,
		Quantity:       int(craftJobRow.Quantity),
		OutputItemID:   int(craftJobRow.OutputItemID),
		OutputQuantity: int(craftJobRow.OutputQuantity),
		Fee:            money.Micro(craftJobRow.FeeMicro),
		Inputs:         inputs,
		Status:         Status(craftJobRow.Status),
		StartedAt:      craftJobRow.StartedAt,
		EndsAt:         craftJobRow.EndsAt,
		DeliveredAt:    craftJobRow.DeliveredAt,
	}, nil
}

func scaledQuantities(quantities []catalog.ItemQuantity, multiplier int) []catalog.ItemQuantity {
	scaled := make([]catalog.ItemQuantity, 0, len(quantities))
	for _, quantity := range quantities {
		scaled = append(scaled, catalog.ItemQuantity{ItemID: quantity.ItemID, Quantity: quantity.Quantity * multiplier})
	}
	return scaled
}
