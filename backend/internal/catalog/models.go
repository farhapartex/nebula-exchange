package catalog

import (
	"encoding/json"

	"nebula-exchange/backend/internal/platform/money"
)

type Category string

const (
	CategoryResource   Category = "resource"
	CategoryComponent  Category = "component"
	CategoryDrill      Category = "drill"
	CategoryShip       Category = "ship"
	CategoryConsumable Category = "consumable"
	CategoryLegendary  Category = "legendary"
)

type Item struct {
	ID            int             `json:"id"`
	Slug          string          `json:"slug"`
	Name          string          `json:"name"`
	Category      Category        `json:"category"`
	Tier          *int            `json:"tier"`
	RarityRank    int             `json:"rarity_rank"`
	IsTradeable   bool            `json:"is_tradeable"`
	IsAuctionOnly bool            `json:"is_auction_only"`
	MaxSupply     *int64          `json:"max_supply"`
	Description   string          `json:"description"`
	Attributes    json.RawMessage `json:"attributes"`
}

type ItemQuantity struct {
	ItemID   int `json:"item_id"`
	Quantity int `json:"quantity"`
}

type Recipe struct {
	ID             string         `json:"id"`
	OutputItemID   int            `json:"output_item_id"`
	OutputQuantity int            `json:"output_quantity"`
	CraftSeconds   int            `json:"craft_seconds"`
	Fee            money.Micro    `json:"fee"`
	Inputs         []ItemQuantity `json:"inputs"`
}

type Upgrade struct {
	ID         string         `json:"id"`
	FromItemID int            `json:"from_item_id"`
	ToItemID   int            `json:"to_item_id"`
	CraftFee   money.Micro    `json:"craft_fee"`
	BuyPrice   *money.Micro   `json:"buy_price"`
	Inputs     []ItemQuantity `json:"inputs"`
}

type LootEntry struct {
	ItemID            int `json:"item_id"`
	MinimumQuantity   int `json:"minimum_quantity"`
	MaximumQuantity   int `json:"maximum_quantity"`
	ChanceBasisPoints int `json:"chance_basis_points"`
}

type Zone struct {
	ID                 string      `json:"id"`
	Name               string      `json:"name"`
	Description        string      `json:"description"`
	DurationSeconds    int         `json:"duration_seconds"`
	FuelCost           int         `json:"fuel_cost"`
	MinimumDrillTier   int         `json:"minimum_drill_tier"`
	AllowedShipItemIDs []int       `json:"allowed_ship_item_ids"`
	Loot               []LootEntry `json:"loot"`
}

type ShopItem struct {
	SKU         string         `json:"sku"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Price       money.Micro    `json:"price"`
	Contents    []ItemQuantity `json:"contents"`
}

type Snapshot struct {
	Items     []Item
	Recipes   []Recipe
	Upgrades  []Upgrade
	Zones     []Zone
	ShopItems []ShopItem
}

func (snapshot Snapshot) ItemByID(itemID int) (Item, bool) {
	for _, item := range snapshot.Items {
		if item.ID == itemID {
			return item, true
		}
	}
	return Item{}, false
}
