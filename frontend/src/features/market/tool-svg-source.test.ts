import { describe, expect, it } from "vitest";

import { toolSvgSource } from "@/features/market/tool-svg-source";

describe("toolSvgSource", () => {
  it("turns SVG markup into an image data URL so it renders as an image, not as page markup", () => {
    const source = toolSvgSource('<svg xmlns="http://www.w3.org/2000/svg"><rect fill="#fff"/></svg>');
    expect(source.startsWith("data:image/svg+xml;charset=utf-8,")).toBe(true);
    expect(source).not.toContain("<");
    expect(decodeURIComponent(source.split(",")[1])).toContain('fill="#fff"');
  });
});
