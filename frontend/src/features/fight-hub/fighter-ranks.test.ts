import { describe, expect, it } from "vitest";

import { rankTitleForLevel } from "@/features/fight-hub/fighter-ranks";

describe("rankTitleForLevel", () => {
  it("starts every fighter as a nobody and grows with level", () => {
    expect(rankTitleForLevel(1)).toBe("Nobody");
    expect(rankTitleForLevel(3)).toBe("Street rat");
    expect(rankTitleForLevel(9)).toBe("Brawler");
    expect(rankTitleForLevel(42)).toBe("Contender");
  });
});
