import { describe, expect, it } from "vitest";

import { calculatePriceChange, formatBasisPointsAsPercent, fractionDigitsForTickSize } from "@/utils/money/price";

describe("fractionDigitsForTickSize", () => {
  it("derives display digits from the market tick size", () => {
    expect(fractionDigitsForTickSize("100")).toBe(4);
    expect(fractionDigitsForTickSize("1000")).toBe(3);
    expect(fractionDigitsForTickSize("10000")).toBe(2);
    expect(fractionDigitsForTickSize("1")).toBe(6);
  });

  it("rejects non-positive tick sizes", () => {
    expect(() => fractionDigitsForTickSize("0")).toThrow();
  });
});

describe("calculatePriceChange", () => {
  it("computes the 24h change in basis points", () => {
    expect(calculatePriceChange("10421", "10000")).toEqual({ direction: "up", changeInBasisPoints: 421n });
    expect(calculatePriceChange("8242", "8420")).toEqual({ direction: "down", changeInBasisPoints: -211n });
    expect(calculatePriceChange("200000", "200000")).toEqual({ direction: "flat", changeInBasisPoints: 0n });
  });

  it("treats a missing reference price as flat", () => {
    expect(calculatePriceChange("100", "0")).toEqual({ direction: "flat", changeInBasisPoints: 0n });
  });
});

describe("formatBasisPointsAsPercent", () => {
  it("formats with sign and two decimals", () => {
    expect(formatBasisPointsAsPercent(421n)).toBe("+4.21%");
    expect(formatBasisPointsAsPercent(-211n)).toBe("−2.11%");
    expect(formatBasisPointsAsPercent(0n)).toBe("0.00%");
    expect(formatBasisPointsAsPercent(12345n)).toBe("+123.45%");
  });
});
