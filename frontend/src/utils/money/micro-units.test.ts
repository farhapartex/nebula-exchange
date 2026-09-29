import { describe, expect, it } from "vitest";

import { formatMicroUnits, parseNcToMicroUnits, toMicroUnits } from "@/utils/money/micro-units";

describe("toMicroUnits", () => {
  it("accepts bigint and integer strings", () => {
    expect(toMicroUnits(5n)).toBe(5n);
    expect(toMicroUnits("9223372036854775807")).toBe(9223372036854775807n);
    expect(toMicroUnits("-42")).toBe(-42n);
  });

  it("rejects non-integer strings", () => {
    expect(() => toMicroUnits("1.5")).toThrow();
    expect(() => toMicroUnits("abc")).toThrow();
  });
});

describe("formatMicroUnits", () => {
  it("formats full precision with grouping by default", () => {
    expect(formatMicroUnits("1284500000")).toBe("1,284.500000");
  });

  it("truncates instead of rounding when showing fewer digits", () => {
    expect(formatMicroUnits("1999999", { fractionDigits: 2 })).toBe("1.99");
    expect(formatMicroUnits("5000", { fractionDigits: 2 })).toBe("0.00");
  });

  it("keeps exact values beyond the safe JavaScript number range", () => {
    expect(formatMicroUnits("9223372036854775807", { fractionDigits: 6 })).toBe("9,223,372,036,854.775807");
  });

  it("handles signs", () => {
    expect(formatMicroUnits("-2500000", { fractionDigits: 2 })).toBe("−2.50");
    expect(formatMicroUnits("2500000", { fractionDigits: 2, showPlusSign: true })).toBe("+2.50");
    expect(formatMicroUnits("-4000", { fractionDigits: 2 })).toBe("0.00");
  });

  it("supports whole numbers without grouping", () => {
    expect(formatMicroUnits("12000000000", { fractionDigits: 0, useGrouping: false })).toBe("12000");
  });
});

describe("parseNcToMicroUnits", () => {
  it("parses decimal NC input", () => {
    expect(parseNcToMicroUnits("0.0105")).toBe(10500n);
    expect(parseNcToMicroUnits("1,284.5")).toBe(1284500000n);
    expect(parseNcToMicroUnits(".5")).toBe(500000n);
    expect(parseNcToMicroUnits("7")).toBe(7000000n);
  });

  it("rejects invalid or too precise input", () => {
    expect(parseNcToMicroUnits("")).toBeNull();
    expect(parseNcToMicroUnits(".")).toBeNull();
    expect(parseNcToMicroUnits("1.2345678")).toBeNull();
    expect(parseNcToMicroUnits("-1")).toBeNull();
    expect(parseNcToMicroUnits("1e5")).toBeNull();
  });
});
