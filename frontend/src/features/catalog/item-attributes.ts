import type { CatalogItem } from "@/features/catalog/api/catalog-types";

export type ItemAttribute = { label: string; value: string };

function describeSpeed(speedMultiplier: string): string {
  const multiplier = Number(speedMultiplier);
  if (multiplier === 1) {
    return "Standard";
  }
  const percentChange = Math.round((1 - multiplier) * 100);
  return percentChange > 0 ? `${percentChange}% faster` : `${-percentChange}% slower`;
}

export function listItemAttributes(catalogItem: CatalogItem): ItemAttribute[] {
  const itemAttributes: ItemAttribute[] = [];
  if (catalogItem.attributes.drill_multiplier) {
    itemAttributes.push({ label: "Yield", value: `×${catalogItem.attributes.drill_multiplier}` });
  }
  if (catalogItem.attributes.cargo_capacity !== undefined) {
    itemAttributes.push({ label: "Cargo", value: `${catalogItem.attributes.cargo_capacity} units` });
  }
  if (catalogItem.attributes.speed_multiplier) {
    itemAttributes.push({ label: "Speed", value: describeSpeed(catalogItem.attributes.speed_multiplier) });
  }
  if (catalogItem.max_supply !== null) {
    itemAttributes.push({ label: "Supply", value: `${catalogItem.max_supply} ever` });
  }
  return itemAttributes;
}
