import { describe, expect, it } from "vitest";

import { formatCoins } from "@/utils/money/format-coins";

describe("formatCoins", () => {
  it("shows whole coins from micro-units without floating point", () => {
    expect(formatCoins("499000000")).toBe("499");
    expect(formatCoins("2400000000")).toBe("2,400");
    expect(formatCoins("0")).toBe("0");
  });

  it("keeps exact fractions and negative balances", () => {
    expect(formatCoins("1500000")).toBe("1.5");
    expect(formatCoins("-499000000")).toBe("-499");
    expect(formatCoins("9007199254740993000000")).toBe("9,007,199,254,740,993");
  });
});
