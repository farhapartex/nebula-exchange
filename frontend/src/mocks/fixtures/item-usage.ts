import type { ItemQuantity, ItemUsage } from "@/features/catalog/api/catalog-types";
import { mockRecipes, mockShopItems, mockUpgrades, mockZones } from "@/mocks/fixtures/catalog-fixtures";

function containsItem(quantities: ItemQuantity[], itemID: number): boolean {
  return quantities.some((quantity) => quantity.item_id === itemID);
}

export function buildMockItemUsage(itemID: number): ItemUsage {
  return {
    input_to_recipes: mockRecipes.filter((recipe) => containsItem(recipe.inputs, itemID)).map((recipe) => recipe.id),
    input_to_upgrades: mockUpgrades
      .filter((upgrade) => containsItem(upgrade.inputs, itemID))
      .map((upgrade) => upgrade.id),
    crafted_by: mockRecipes.filter((recipe) => recipe.output_item_id === itemID).map((recipe) => recipe.id),
    upgraded_from: mockUpgrades.filter((upgrade) => upgrade.to_item_id === itemID).map((upgrade) => upgrade.id),
    upgrades_into: mockUpgrades.filter((upgrade) => upgrade.from_item_id === itemID).map((upgrade) => upgrade.id),
    dropped_in_zones: mockZones
      .filter((zone) => zone.loot.some((lootEntry) => lootEntry.item_id === itemID))
      .map((zone) => zone.id),
    sold_as_skus: mockShopItems
      .filter((shopItem) => containsItem(shopItem.contents, itemID))
      .map((shopItem) => shopItem.sku),
  };
}
