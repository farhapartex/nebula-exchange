import { describe, expect, it } from "vitest";

import { describeLevelsStillNeeded, levelsStillNeeded } from "@/features/market/fighter-level-requirement";

describe("levelsStillNeeded", () => {
  it("counts the fighter levels a player still has to earn", () => {
    expect(levelsStillNeeded(3, 1)).toBe(2);
    expect(levelsStillNeeded(3, 3)).toBe(0);
    expect(levelsStillNeeded(2, 5)).toBe(0);
    expect(levelsStillNeeded(4, undefined)).toBe(0);
  });

  it("describes the gap in plain words", () => {
    expect(describeLevelsStillNeeded(1)).toBe("Earn 1 more level");
    expect(describeLevelsStillNeeded(3)).toBe("Earn 3 more levels");
  });
});
