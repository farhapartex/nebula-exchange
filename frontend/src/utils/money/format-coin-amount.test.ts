import { describe, expect, it } from "vitest";

import { formatCoinAmount } from "@/utils/money/format-coin-amount";

describe("formatCoinAmount", () => {
  it("groups whole coins with commas", () => {
    expect(formatCoinAmount("300")).toBe("300");
    expect(formatCoinAmount("1250")).toBe("1,250");
    expect(formatCoinAmount(1_000_000n)).toBe("1,000,000");
  });
});
