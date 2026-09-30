import type { CatalogItem, LootEntry, Zone } from "@/features/catalog/api/catalog-types";
import { applyBasisPoints, basisPointsPerWhole, parseBasisPoints } from "@/utils/numbers/basis-points";

export const maximumActiveMissions = 3;
export const fuelCellItemID = 401;
const legendaryDrillTier = 5;

export type DrillProfile = { tier: number; yieldBasisPoints: number };
export type ShipProfile = { cargoCapacity: number; speedBasisPoints: number };

export function drillProfileOf(catalogItem: CatalogItem): DrillProfile | null {
  const yieldBasisPoints = parseBasisPoints(catalogItem.attributes.drill_multiplier);
  if (yieldBasisPoints === null || (catalogItem.category !== "drill" && catalogItem.category !== "legendary")) {
    return null;
  }
  return { tier: catalogItem.tier ?? legendaryDrillTier, yieldBasisPoints };
}

export function shipProfileOf(catalogItem: CatalogItem): ShipProfile | null {
  const speedBasisPoints = parseBasisPoints(catalogItem.attributes.speed_multiplier);
  const cargoCapacity = catalogItem.attributes.cargo_capacity;
  if (catalogItem.category !== "ship" || speedBasisPoints === null || !cargoCapacity) {
    return null;
  }
  return { cargoCapacity, speedBasisPoints };
}

export function zoneAllowsShip(zone: Zone, shipItemID: number): boolean {
  return zone.allowed_ship_item_ids.length === 0 || zone.allowed_ship_item_ids.includes(shipItemID);
}

export function missionDurationSeconds(zone: Zone, ship: ShipProfile): number {
  return applyBasisPoints(zone.duration_seconds, ship.speedBasisPoints);
}

export type LootRange = { itemID: number; minimum: number; maximum: number; chanceBasisPoints: number };

export function expectedLootRanges(loot: LootEntry[], drill: DrillProfile | null): LootRange[] {
  return loot.map((lootEntry) => {
    const isChanceDrop = lootEntry.chance_basis_points < basisPointsPerWhole;
    const multiplier = drill && !isChanceDrop ? drill.yieldBasisPoints : basisPointsPerWhole;
    return {
      itemID: lootEntry.item_id,
      minimum: applyBasisPoints(lootEntry.minimum_quantity, multiplier),
      maximum: applyBasisPoints(lootEntry.maximum_quantity, multiplier),
      chanceBasisPoints: lootEntry.chance_basis_points,
    };
  });
}

export function maximumLootUnits(ranges: LootRange[]): number {
  return ranges.reduce((totalUnits, range) => totalUnits + range.maximum, 0);
}

export function describeZoneRequirements(zone: Zone, itemsByID: Map<number, CatalogItem>): string[] {
  const requirements: string[] = [];
  if (zone.minimum_drill_tier > 1) {
    requirements.push(`Drill T${zone.minimum_drill_tier}+`);
  }
  if (zone.allowed_ship_item_ids.length > 0) {
    requirements.push(
      zone.allowed_ship_item_ids.map((shipItemID) => itemsByID.get(shipItemID)?.name ?? "Ship").join(" / "),
    );
  }
  return requirements;
}
