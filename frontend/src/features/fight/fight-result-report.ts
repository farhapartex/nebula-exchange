import type { FightResultReport } from "@/features/fight/api/fight-session-api";
import type { FightSnapshot } from "@/features/fight/engine/fight-controller";

export function buildFightResultReport(snapshot: FightSnapshot): FightResultReport | null {
  if (!snapshot.outcome) {
    return null;
  }
  return {
    outcome: snapshot.outcome === "won" ? "WON" : "LOST",
    duration_ms: Math.max(snapshot.elapsedMs, 1),
    damage_dealt: snapshot.damageDealt,
    damage_taken: snapshot.damageTaken,
  };
}
