import { describe, expect, it } from "vitest";

import { safeHexColor } from "@/utils/colors/safe-hex-color";

describe("safeHexColor", () => {
  it("accepts 3, 4, 6 and 8 digit hex colors", () => {
    expect(safeHexColor("#fff", "#000")).toBe("#fff");
    expect(safeHexColor("#2dd4bf26", "#000")).toBe("#2dd4bf26");
  });

  it("falls back for anything that is not a plain hex color", () => {
    expect(safeHexColor("red", "#000")).toBe("#000");
    expect(safeHexColor("url(https://evil.example)", "#000")).toBe("#000");
    expect(safeHexColor(null, "#000")).toBe("#000");
  });
});
