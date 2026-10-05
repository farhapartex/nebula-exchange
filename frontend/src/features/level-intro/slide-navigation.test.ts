import { describe, expect, it } from "vitest";

import { isFinalSlide, nextSlideIndex, previousSlideIndex } from "@/features/level-intro/slide-navigation";

describe("slide navigation", () => {
  it("never moves past either end", () => {
    expect(nextSlideIndex(0, 6)).toBe(1);
    expect(nextSlideIndex(5, 6)).toBe(5);
    expect(previousSlideIndex(0)).toBe(0);
    expect(previousSlideIndex(3)).toBe(2);
  });

  it("knows when the call to action is showing", () => {
    expect(isFinalSlide(4, 6)).toBe(false);
    expect(isFinalSlide(5, 6)).toBe(true);
  });
});
