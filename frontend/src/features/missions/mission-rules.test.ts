import { describe, expect, it } from "vitest";

import {
  drillProfileOf,
  expectedLootRanges,
  maximumLootUnits,
  missionDurationSeconds,
  shipProfileOf,
  zoneAllowsShip,
} from "@/features/missions/mission-rules";
import { mockCatalogItems, mockZones } from "@/mocks/fixtures/catalog-fixtures";

function itemBySlug(slug: string) {
  const catalogItem = mockCatalogItems.find((candidate) => candidate.slug === slug);
  if (!catalogItem) {
    throw new Error(slug);
  }
  return catalogItem;
}

function zoneByID(zoneID: string) {
  const zone = mockZones.find((candidate) => candidate.id === zoneID);
  if (!zone) {
    throw new Error(zoneID);
  }
  return zone;
}

describe("mission rules", () => {
  it("reads drill and ship profiles from catalog attributes", () => {
    expect(drillProfileOf(itemBySlug("drill-t2"))).toEqual({ tier: 2, yieldBasisPoints: 12_500 });
    expect(drillProfileOf(itemBySlug("relic-drill-of-orion"))).toEqual({ tier: 5, yieldBasisPoints: 30_000 });
    expect(drillProfileOf(itemBySlug("scout"))).toBeNull();
    expect(shipProfileOf(itemBySlug("interceptor"))).toEqual({ cargoCapacity: 40, speedBasisPoints: 6_000 });
  });

  it("matches the server's duration and loot math", () => {
    const interceptor = shipProfileOf(itemBySlug("interceptor"));
    expect(interceptor && missionDurationSeconds(zoneByID("asteroid-belt"), interceptor)).toBe(540);

    const ranges = expectedLootRanges(zoneByID("deep-void").loot, drillProfileOf(itemBySlug("drill-t4")));
    expect(ranges).toEqual([
      { itemID: 5, minimum: 4, maximum: 8, chanceBasisPoints: 10_000 },
      { itemID: 6, minimum: 1, maximum: 1, chanceBasisPoints: 500 },
    ]);
    expect(maximumLootUnits(ranges)).toBe(9);
  });

  it("knows which ships each zone accepts", () => {
    expect(zoneAllowsShip(zoneByID("asteroid-belt"), 301)).toBe(true);
    expect(zoneAllowsShip(zoneByID("plasma-nebula"), 301)).toBe(false);
    expect(zoneAllowsShip(zoneByID("deep-void"), 304)).toBe(true);
  });
});
