package missions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/ledger/ledgerstore"
)

const inventoryScanLimit = 1000

type ZoneAvailability struct {
	pool    *pgxpool.Pool
	catalog *catalog.Service
}

func NewZoneAvailability(pool *pgxpool.Pool, catalogService *catalog.Service) *ZoneAvailability {
	return &ZoneAvailability{pool: pool, catalog: catalogService}
}

func (availability *ZoneAvailability) ForPlayer(ctx context.Context, userID uuid.UUID, zones []catalog.Zone) (map[string]catalog.ZoneUnlock, error) {
	catalogSnapshot, err := availability.catalog.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	holdingRows, err := ledgerstore.New(availability.pool).ListPlayerInventory(ctx, ledgerstore.ListPlayerInventoryParams{
		UserID: &userID, AfterItemID: 0, RowLimit: inventoryScanLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("load inventory for zones: %w", err)
	}

	bestDrillTier := 0
	availableShips := map[int]bool{}
	availableFuel := int64(0)
	for _, holdingRow := range holdingRows {
		if holdingRow.Available <= 0 {
			continue
		}
		itemID := int(holdingRow.ItemID)
		if itemID == fuelCellItemID {
			availableFuel = holdingRow.Available
		}
		heldItem, _ := catalogSnapshot.ItemByID(itemID)
		if drillStats, isDrill := heldItem.DrillStats(); isDrill {
			bestDrillTier = max(bestDrillTier, drillStats.Tier)
		}
		if _, isShip := heldItem.ShipStats(); isShip {
			availableShips[itemID] = true
		}
	}

	unlocks := make(map[string]catalog.ZoneUnlock, len(zones))
	for _, zone := range zones {
		hasAllowedShip := false
		for shipItemID := range availableShips {
			if zone.AllowsShip(shipItemID) {
				hasAllowedShip = true
				break
			}
		}
		unlock := catalog.ZoneUnlock{
			HasRequiredDrill: bestDrillTier >= zone.MinimumDrillTier,
			HasAllowedShip:   hasAllowedShip,
			HasEnoughFuel:    availableFuel >= int64(zone.FuelCost),
		}
		unlock.IsUnlocked = unlock.HasRequiredDrill && unlock.HasAllowedShip && unlock.HasEnoughFuel
		unlocks[zone.ID] = unlock
	}
	return unlocks, nil
}
