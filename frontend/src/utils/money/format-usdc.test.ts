import { describe, expect, it } from "vitest";

import { formatUsdcFromCents } from "@/utils/money/format-usdc";

describe("formatUsdcFromCents", () => {
  it("shows a USD cent price as the same USDC amount", () => {
    expect(formatUsdcFromCents(499n)).toBe("4.99 USDC");
    expect(formatUsdcFromCents(159700n)).toBe("1,597.00 USDC");
    expect(formatUsdcFromCents(5n)).toBe("0.05 USDC");
  });
});
