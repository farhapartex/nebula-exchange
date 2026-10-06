import { describe, expect, it } from "vitest";

import type { FightSnapshot } from "@/features/fight/engine/fight-controller";
import { buildFightResultReport } from "@/features/fight/fight-result-report";

const fighter = { health: 0, maxHealth: 100, stamina: 50, maxStamina: 100 };

function snapshotWith(overrides: Partial<FightSnapshot>): FightSnapshot {
  return {
    player: fighter,
    enemy: fighter,
    remainingSeconds: 50,
    elapsedMs: 40_250,
    outcome: null,
    isPaused: false,
    damageDealt: 96,
    damageTaken: 30,
    ...overrides,
  };
}

describe("buildFightResultReport", () => {
  it("reports the fight time and damage once the fight has an outcome", () => {
    expect(buildFightResultReport(snapshotWith({ outcome: "won" }))).toEqual({
      outcome: "WON",
      duration_ms: 40_250,
      damage_dealt: 96,
      damage_taken: 30,
    });
    expect(buildFightResultReport(snapshotWith({ outcome: "lost" }))?.outcome).toBe("LOST");
  });

  it("reports nothing while the fight is still going", () => {
    expect(buildFightResultReport(snapshotWith({ outcome: null }))).toBeNull();
  });
});
