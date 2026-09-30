import { describe, expect, it } from "vitest";

import { applyBasisPoints, parseBasisPoints } from "@/utils/numbers/basis-points";

describe("basis points", () => {
  it("parses decimal multipliers exactly", () => {
    expect(parseBasisPoints("1.25")).toBe(12_500);
    expect(parseBasisPoints("0.60")).toBe(6_000);
    expect(parseBasisPoints("3")).toBe(30_000);
    expect(parseBasisPoints("abc")).toBeNull();
    expect(parseBasisPoints(undefined)).toBeNull();
  });

  it("applies multipliers and rounds down like the server", () => {
    expect(applyBasisPoints(14, 12_500)).toBe(17);
    expect(applyBasisPoints(900, 6_000)).toBe(540);
  });
});
