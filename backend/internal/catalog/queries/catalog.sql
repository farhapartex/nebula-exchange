-- name: ListItems :many
SELECT id, slug, name, category, tier, rarity_rank, is_tradeable, is_auction_only, max_supply, description, attributes
FROM items
ORDER BY id;

-- name: ListRecipes :many
SELECT id, output_item_id, output_quantity, craft_seconds, fee_micro
FROM recipes
WHERE is_enabled = true
ORDER BY output_item_id;

-- name: ListRecipeInputs :many
SELECT recipe_id, item_id, quantity
FROM recipe_inputs
ORDER BY recipe_id, item_id;

-- name: ListUpgrades :many
SELECT id, from_item_id, to_item_id, craft_fee_micro, buy_price_micro
FROM upgrades
WHERE is_enabled = true
ORDER BY from_item_id, to_item_id;

-- name: ListUpgradeInputs :many
SELECT upgrade_id, item_id, quantity
FROM upgrade_inputs
ORDER BY upgrade_id, item_id;

-- name: ListZones :many
SELECT id, name, description, duration_seconds, fuel_cost, minimum_drill_tier, allowed_ship_item_ids
FROM zones
WHERE is_enabled = true
ORDER BY sort_order;

-- name: ListLootTables :many
SELECT zone_id, item_id, minimum_quantity, maximum_quantity, chance_basis_points
FROM loot_tables
ORDER BY zone_id, item_id;

-- name: ListShopSkus :many
SELECT sku, name, description, price_micro
FROM shop_skus
WHERE is_enabled = true
ORDER BY sort_order;

-- name: ListShopSkuItems :many
SELECT sku, item_id, quantity
FROM shop_sku_items
ORDER BY sku, item_id;
