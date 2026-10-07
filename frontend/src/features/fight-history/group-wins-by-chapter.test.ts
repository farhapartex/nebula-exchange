import { describe, expect, it } from "vitest";

import type { WonFight } from "@/features/fight-history/api/fight-history-api";
import { groupWinsByChapter } from "@/features/fight-history/group-wins-by-chapter";

function winIn(chapterNumber: number, fightID: string): WonFight {
  return {
    id: fightID,
    level: {
      id: `${chapterNumber}-1`,
      number: 1,
      title: "x",
      chapter_number: chapterNumber,
      chapter_title: `Chapter ${chapterNumber}`,
    },
    stars: 3,
    duration_ms: 40000,
    damage_dealt: 90,
    damage_taken: 20,
    finished_at: "2026-10-07T10:00:00Z",
  };
}

describe("groupWinsByChapter", () => {
  it("keeps the order and puts wins of the same chapter together", () => {
    const groups = groupWinsByChapter([winIn(1, "a"), winIn(1, "b"), winIn(2, "c")]);
    expect(groups.map((group) => [group.chapterNumber, group.wins.map((win) => win.id)])).toEqual([
      [1, ["a", "b"]],
      [2, ["c"]],
    ]);
  });

  it("returns no groups when there are no wins", () => {
    expect(groupWinsByChapter([])).toEqual([]);
  });
});
