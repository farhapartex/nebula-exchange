import { describe, expect, it } from "vitest";

import type { StoryLevel } from "@/features/fight-hub/api/fight-hub-types";
import { markLevelsByProgress } from "@/features/fight-hub/story-progress";

function chapterLevels(): StoryLevel[] {
  const levels: StoryLevel[] = [1, 2, 3].map((levelNumber) => ({
    id: `1-${levelNumber}`,
    number: levelNumber,
    title: `Level ${levelNumber}`,
    teaser: "",
    status: "locked",
    best_stars: null,
  }));
  return [...levels, { id: "finale", number: null, title: "Finale", teaser: "", status: "locked", best_stars: null }];
}

describe("markLevelsByProgress", () => {
  it("marks won levels completed and the next one current", () => {
    expect(markLevelsByProgress(chapterLevels(), 1).map((level) => level.status)).toEqual([
      "completed",
      "current",
      "locked",
      "locked",
    ]);
  });

  it("starts a new player on level 1", () => {
    expect(markLevelsByProgress(chapterLevels(), 0)[0].status).toBe("current");
  });

  it("opens the finale once every level is won", () => {
    expect(markLevelsByProgress(chapterLevels(), 3).map((level) => level.status)).toEqual([
      "completed",
      "completed",
      "completed",
      "current",
    ]);
  });
});
