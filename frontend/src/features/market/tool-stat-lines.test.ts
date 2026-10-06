import { describe, expect, it } from "vitest";

import { toolStatLines } from "@/features/market/tool-stat-lines";

describe("toolStatLines", () => {
  it("turns tool stats into short readable lines", () => {
    expect(toolStatLines({ damage_bonus: 4, reach_bonus: 18, swing_speed_modifier: -0.05 })).toEqual([
      "+4 damage",
      "+18 reach",
      "5% slower swings",
    ]);
    expect(toolStatLines({ block_damage_reduction: 0.25 })).toEqual(["25% less damage when blocking"]);
  });

  it("still shows stats it does not know yet", () => {
    expect(toolStatLines({ bleed_chance: 0.1 })).toEqual(["Bleed chance: 0.1"]);
  });
});
