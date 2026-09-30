import { describe, expect, it } from "vitest";

import { formatLootChance, formatQuantityRange } from "@/features/catalog/loot-chance";

describe("loot formatting", () => {
  it("describes drop chances in player terms", () => {
    expect(formatLootChance(10_000)).toBe("Always");
    expect(formatLootChance(500)).toBe("5% chance");
    expect(formatLootChance(25)).toBe("0.25% chance");
  });

  it("collapses equal quantity ranges", () => {
    expect(formatQuantityRange(1, 1)).toBe("1");
    expect(formatQuantityRange(8, 14)).toBe("8–14");
  });
});
