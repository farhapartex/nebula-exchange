import { describe, expect, it } from "vitest";

import { formatTokenAmount } from "@/utils/web3/format-token-amount";

describe("formatTokenAmount", () => {
  it("shows a token amount with a limited number of decimals, rounding up so it is never understated", () => {
    expect(formatTokenAmount(1_851_854_398_368_133n, 18, 6)).toBe("0.001852");
    expect(formatTokenAmount(4_990_000n, 6, 2)).toBe("4.99");
    expect(formatTokenAmount(10_000_000_000_000_000_000_000n, 18, 6)).toBe("10000");
    expect(formatTokenAmount(1_500_000_000_000_000_000n, 18, 6)).toBe("1.5");
  });
});
