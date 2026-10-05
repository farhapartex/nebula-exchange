import { describe, expect, it } from "vitest";

import {
  advancePacing,
  createFightPacing,
  freezeForImpact,
  isInSlowMotion,
  startSlowMotion,
} from "@/features/fight/stage/fight-pacing";

describe("fight pacing", () => {
  it("passes real time through when nothing is happening", () => {
    expect(advancePacing(createFightPacing(), 16)).toBe(16);
  });

  it("freezes time during a hit stop and then resumes", () => {
    const pacing = createFightPacing();
    freezeForImpact(pacing, 30);
    expect(advancePacing(pacing, 16)).toBe(0);
    expect(advancePacing(pacing, 16)).toBe(0);
    expect(advancePacing(pacing, 16)).toBe(16);
  });

  it("slows time after a knockout until the slow motion ends", () => {
    const pacing = createFightPacing();
    startSlowMotion(pacing, 20);
    expect(advancePacing(pacing, 16)).toBe(4);
    expect(isInSlowMotion(pacing)).toBe(true);
    advancePacing(pacing, 16);
    expect(isInSlowMotion(pacing)).toBe(false);
  });
});
