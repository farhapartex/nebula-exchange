import { describe, expect, it } from "vitest";

import { hasAllRequirements, maximumAffordableBatches } from "@/features/workshop/requirement-rules";

describe("workshop requirements", () => {
  const alloyInputs = [
    { item_id: 1, quantity: 5 },
    { item_id: 2, quantity: 2 },
  ];

  it("checks owned quantities against a batch", () => {
    const owned = new Map([
      [1, 11],
      [2, 4],
    ]);
    expect(hasAllRequirements(alloyInputs, owned, 2)).toBe(true);
    expect(hasAllRequirements(alloyInputs, owned, 3)).toBe(false);
    expect(maximumAffordableBatches(alloyInputs, owned)).toBe(2);
  });

  it("treats missing items as zero", () => {
    expect(maximumAffordableBatches(alloyInputs, new Map([[1, 50]]))).toBe(0);
  });
});
