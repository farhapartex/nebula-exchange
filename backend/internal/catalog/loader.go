package catalog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/catalog/catalogstore"
	"nebula-exchange/backend/internal/platform/money"
)

type Loader struct {
	queries *catalogstore.Queries
}

func NewLoader(pool *pgxpool.Pool) *Loader {
	return &Loader{queries: catalogstore.New(pool)}
}

func (loader *Loader) Load(ctx context.Context) (Snapshot, error) {
	items, err := loader.loadItems(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	recipes, err := loader.loadRecipes(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	upgrades, err := loader.loadUpgrades(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	zones, err := loader.loadZones(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	shopItems, err := loader.loadShopItems(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Items: items, Recipes: recipes, Upgrades: upgrades, Zones: zones, ShopItems: shopItems}, nil
}

func (loader *Loader) loadItems(ctx context.Context) ([]Item, error) {
	itemRows, err := loader.queries.ListItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	items := make([]Item, 0, len(itemRows))
	for _, itemRow := range itemRows {
		item := Item{
			ID:            int(itemRow.ID),
			Slug:          itemRow.Slug,
			Name:          itemRow.Name,
			Category:      Category(itemRow.Category),
			RarityRank:    int(itemRow.RarityRank),
			IsTradeable:   itemRow.IsTradeable,
			IsAuctionOnly: itemRow.IsAuctionOnly,
			Description:   itemRow.Description,
			Attributes:    json.RawMessage(itemRow.Attributes),
		}
		if itemRow.Tier.Valid {
			tier := int(itemRow.Tier.Int32)
			item.Tier = &tier
		}
		if itemRow.MaxSupply.Valid {
			maxSupply := itemRow.MaxSupply.Int64
			item.MaxSupply = &maxSupply
		}
		items = append(items, item)
	}
	return items, nil
}

func (loader *Loader) loadRecipes(ctx context.Context) ([]Recipe, error) {
	recipeRows, err := loader.queries.ListRecipes(ctx)
	if err != nil {
		return nil, fmt.Errorf("list recipes: %w", err)
	}
	inputRows, err := loader.queries.ListRecipeInputs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list recipe inputs: %w", err)
	}
	inputsByRecipe := map[string][]ItemQuantity{}
	for _, inputRow := range inputRows {
		inputsByRecipe[inputRow.RecipeID] = append(inputsByRecipe[inputRow.RecipeID], ItemQuantity{ItemID: int(inputRow.ItemID), Quantity: int(inputRow.Quantity)})
	}
	recipes := make([]Recipe, 0, len(recipeRows))
	for _, recipeRow := range recipeRows {
		recipes = append(recipes, Recipe{
			ID:             recipeRow.ID,
			OutputItemID:   int(recipeRow.OutputItemID),
			OutputQuantity: int(recipeRow.OutputQuantity),
			CraftSeconds:   int(recipeRow.CraftSeconds),
			Fee:            money.Micro(recipeRow.FeeMicro),
			Inputs:         nonNilQuantities(inputsByRecipe[recipeRow.ID]),
		})
	}
	return recipes, nil
}

func (loader *Loader) loadUpgrades(ctx context.Context) ([]Upgrade, error) {
	upgradeRows, err := loader.queries.ListUpgrades(ctx)
	if err != nil {
		return nil, fmt.Errorf("list upgrades: %w", err)
	}
	inputRows, err := loader.queries.ListUpgradeInputs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list upgrade inputs: %w", err)
	}
	inputsByUpgrade := map[string][]ItemQuantity{}
	for _, inputRow := range inputRows {
		inputsByUpgrade[inputRow.UpgradeID] = append(inputsByUpgrade[inputRow.UpgradeID], ItemQuantity{ItemID: int(inputRow.ItemID), Quantity: int(inputRow.Quantity)})
	}
	upgrades := make([]Upgrade, 0, len(upgradeRows))
	for _, upgradeRow := range upgradeRows {
		upgrade := Upgrade{
			ID:         upgradeRow.ID,
			FromItemID: int(upgradeRow.FromItemID),
			ToItemID:   int(upgradeRow.ToItemID),
			CraftFee:   money.Micro(upgradeRow.CraftFeeMicro),
			Inputs:     nonNilQuantities(inputsByUpgrade[upgradeRow.ID]),
		}
		if upgradeRow.BuyPriceMicro.Valid {
			buyPrice := money.Micro(upgradeRow.BuyPriceMicro.Int64)
			upgrade.BuyPrice = &buyPrice
		}
		upgrades = append(upgrades, upgrade)
	}
	return upgrades, nil
}

func (loader *Loader) loadZones(ctx context.Context) ([]Zone, error) {
	zoneRows, err := loader.queries.ListZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("list zones: %w", err)
	}
	lootRows, err := loader.queries.ListLootTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("list loot tables: %w", err)
	}
	lootByZone := map[string][]LootEntry{}
	for _, lootRow := range lootRows {
		lootByZone[lootRow.ZoneID] = append(lootByZone[lootRow.ZoneID], LootEntry{
			ItemID:            int(lootRow.ItemID),
			MinimumQuantity:   int(lootRow.MinimumQuantity),
			MaximumQuantity:   int(lootRow.MaximumQuantity),
			ChanceBasisPoints: int(lootRow.ChanceBasisPoints),
		})
	}
	zones := make([]Zone, 0, len(zoneRows))
	for _, zoneRow := range zoneRows {
		allowedShipItemIDs := make([]int, 0, len(zoneRow.AllowedShipItemIds))
		for _, shipItemID := range zoneRow.AllowedShipItemIds {
			allowedShipItemIDs = append(allowedShipItemIDs, int(shipItemID))
		}
		loot := lootByZone[zoneRow.ID]
		if loot == nil {
			loot = []LootEntry{}
		}
		zones = append(zones, Zone{
			ID:                 zoneRow.ID,
			Name:               zoneRow.Name,
			Description:        zoneRow.Description,
			DurationSeconds:    int(zoneRow.DurationSeconds),
			FuelCost:           int(zoneRow.FuelCost),
			MinimumDrillTier:   int(zoneRow.MinimumDrillTier),
			AllowedShipItemIDs: allowedShipItemIDs,
			Loot:               loot,
		})
	}
	return zones, nil
}

func (loader *Loader) loadShopItems(ctx context.Context) ([]ShopItem, error) {
	skuRows, err := loader.queries.ListShopSkus(ctx)
	if err != nil {
		return nil, fmt.Errorf("list shop skus: %w", err)
	}
	contentRows, err := loader.queries.ListShopSkuItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("list shop sku items: %w", err)
	}
	contentsBySKU := map[string][]ItemQuantity{}
	for _, contentRow := range contentRows {
		contentsBySKU[contentRow.Sku] = append(contentsBySKU[contentRow.Sku], ItemQuantity{ItemID: int(contentRow.ItemID), Quantity: int(contentRow.Quantity)})
	}
	shopItems := make([]ShopItem, 0, len(skuRows))
	for _, skuRow := range skuRows {
		shopItems = append(shopItems, ShopItem{
			SKU:         skuRow.Sku,
			Name:        skuRow.Name,
			Description: skuRow.Description,
			Price:       money.Micro(skuRow.PriceMicro),
			Contents:    nonNilQuantities(contentsBySKU[skuRow.Sku]),
		})
	}
	return shopItems, nil
}

func nonNilQuantities(quantities []ItemQuantity) []ItemQuantity {
	if quantities == nil {
		return []ItemQuantity{}
	}
	return quantities
}
