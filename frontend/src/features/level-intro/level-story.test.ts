import { describe, expect, it } from "vitest";

import type { StorySlide } from "@/features/level-intro/api/level-story-api";
import { arrangeLevelStory } from "@/features/level-intro/level-story";

const palette = { background: "#08090a", glow: "#2dd4bf26", accent: "#2dd4bf" };

function storySlide(position: number, kind: StorySlide["kind"] = "SLIDE"): StorySlide {
  return {
    id: `slide-${position}`,
    position,
    kind,
    eyebrow: null,
    heading: `Heading ${position}`,
    body: "Body",
    image: null,
    palette,
    button_label: kind === "CALL_TO_ACTION" ? "Play" : null,
  };
}

describe("arrangeLevelStory", () => {
  it("orders the slides by position and separates the call to action", () => {
    const levelStory = arrangeLevelStory([storySlide(3, "CALL_TO_ACTION"), storySlide(2), storySlide(1)]);
    expect(levelStory?.slides.map((slide) => slide.position)).toEqual([1, 2]);
    expect(levelStory?.callToAction.button_label).toBe("Play");
  });

  it("treats a story without slides or without a call to action as no story", () => {
    expect(arrangeLevelStory([])).toBeNull();
    expect(arrangeLevelStory([storySlide(1), storySlide(2)])).toBeNull();
    expect(arrangeLevelStory([storySlide(1, "CALL_TO_ACTION")])).toBeNull();
  });
});
