import { describe, expect, it } from "vitest";

import { listItemAttributes } from "@/features/catalog/item-attributes";
import { mockCatalogItems } from "@/mocks/fixtures/catalog-fixtures";

function catalogItem(slug: string) {
  const foundItem = mockCatalogItems.find((candidate) => candidate.slug === slug);
  if (!foundItem) {
    throw new Error(`fixture ${slug} is missing`);
  }
  return foundItem;
}

describe("listItemAttributes", () => {
  it("describes ship cargo and speed in player terms", () => {
    expect(listItemAttributes(catalogItem("interceptor"))).toEqual([
      { label: "Cargo", value: "40 units" },
      { label: "Speed", value: "40% faster" },
    ]);
    expect(listItemAttributes(catalogItem("freighter"))).toContainEqual({ label: "Speed", value: "20% slower" });
    expect(listItemAttributes(catalogItem("scout"))).toContainEqual({ label: "Speed", value: "Standard" });
  });

  it("shows drill yield and legendary supply caps", () => {
    expect(listItemAttributes(catalogItem("drill-t3"))).toEqual([{ label: "Yield", value: "×1.50" }]);
    expect(listItemAttributes(catalogItem("relic-drill-of-orion"))).toEqual([
      { label: "Yield", value: "×3.00" },
      { label: "Supply", value: "1 ever" },
    ]);
  });

  it("has nothing to say about plain resources", () => {
    expect(listItemAttributes(catalogItem("iron-ore"))).toEqual([]);
  });
});
