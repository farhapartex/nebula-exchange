import type { ItemQuantity } from "@/features/catalog/api/catalog-types";

export function hasAllRequirements(
  requirements: ItemQuantity[],
  ownedByItem: Map<number, number>,
  multiplier = 1,
): boolean {
  return requirements.every(
    (requirement) => (ownedByItem.get(requirement.item_id) ?? 0) >= requirement.quantity * multiplier,
  );
}

export function maximumAffordableBatches(requirements: ItemQuantity[], ownedByItem: Map<number, number>): number {
  return Math.min(
    ...requirements.map((requirement) =>
      Math.floor((ownedByItem.get(requirement.item_id) ?? 0) / requirement.quantity),
    ),
  );
}
