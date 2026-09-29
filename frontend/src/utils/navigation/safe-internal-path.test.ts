import { describe, expect, it } from "vitest";

import { safeInternalPath } from "@/utils/navigation/safe-internal-path";

describe("safeInternalPath", () => {
  it("keeps internal paths", () => {
    expect(safeInternalPath("/exchange/IRON-NC")).toBe("/exchange/IRON-NC");
  });

  it("rejects external and protocol-relative redirects", () => {
    expect(safeInternalPath("https://evil.example")).toBeNull();
    expect(safeInternalPath("//evil.example")).toBeNull();
    expect(safeInternalPath("/\\evil.example")).toBeNull();
    expect(safeInternalPath(null)).toBeNull();
  });
});
