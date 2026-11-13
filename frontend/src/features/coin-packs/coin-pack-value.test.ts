import { describe, expect, it } from "vitest";

import type { CoinPack } from "@/features/coin-packs/api/coin-pack-api";
import { cheapestCoinPack, extraCoinsPercent, smallestCoinPackCovering } from "@/features/coin-packs/coin-pack-value";

const starterPack: CoinPack = { id: "coins-500", name: "Pocket change", coins: "500", price_cents: "499" };
const streetPack: CoinPack = { id: "coins-1100", name: "Street stash", coins: "1100", price_cents: "999" };
const bossPack: CoinPack = { id: "coins-2400", name: "Boss bag", coins: "2400", price_cents: "1999" };
const coinPacks = [starterPack, streetPack, bossPack];

describe("extraCoinsPercent", () => {
  it("is 0 for the base pack", () => {
    expect(extraCoinsPercent(starterPack, starterPack)).toBe(0);
  });

  it("rounds the extra coins per dollar against the base pack", () => {
    expect(extraCoinsPercent(streetPack, starterPack)).toBe(10);
    expect(extraCoinsPercent(bossPack, starterPack)).toBe(20);
  });
});

describe("cheapestCoinPack", () => {
  it("picks the lowest price", () => {
    expect(cheapestCoinPack([bossPack, starterPack, streetPack])).toBe(starterPack);
  });

  it("is undefined for no packs", () => {
    expect(cheapestCoinPack([])).toBeUndefined();
  });
});

describe("smallestCoinPackCovering", () => {
  it("picks the cheapest pack that covers the missing coins", () => {
    expect(smallestCoinPackCovering(coinPacks, 200n)).toBe(starterPack);
    expect(smallestCoinPackCovering(coinPacks, 900n)).toBe(streetPack);
  });

  it("falls back to the biggest pack when none covers it", () => {
    expect(smallestCoinPackCovering(coinPacks, 9000n)).toBe(bossPack);
  });
});
