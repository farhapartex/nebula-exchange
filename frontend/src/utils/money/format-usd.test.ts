import { describe, expect, it } from "vitest";

import { formatUsd } from "@/utils/money/format-usd";

describe("formatUsd", () => {
  it("formats whole cents without floating point", () => {
    expect(formatUsd(499n)).toBe("USD 4.99");
    expect(formatUsd(1347n)).toBe("USD 13.47");
    expect(formatUsd(5n)).toBe("USD 0.05");
    expect(formatUsd(123456789n)).toBe("USD 1,234,567.89");
  });

  it("shows a discount as a negative amount", () => {
    expect(formatUsd(-150n)).toBe("-USD 1.50");
  });
});
