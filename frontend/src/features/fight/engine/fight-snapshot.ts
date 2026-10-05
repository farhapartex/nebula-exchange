import type { FightSnapshot } from "@/features/fight/engine/fight-controller";
import type { FightRules } from "@/features/fight/engine/fight-simulation";
import type { FightState } from "@/features/fight/engine/fight-types";

export function buildSnapshot(
  state: FightState,
  rules: FightRules,
  isPaused: boolean,
  damageDealt: number,
  damageTaken: number,
): FightSnapshot {
  return {
    player: {
      health: state.player.health,
      maxHealth: rules.player.max_health,
      stamina: state.player.stamina,
      maxStamina: rules.player.max_stamina,
    },
    enemy: {
      health: state.enemy.health,
      maxHealth: rules.enemy.max_health,
      stamina: state.enemy.stamina,
      maxStamina: rules.enemy.max_stamina,
    },
    remainingSeconds: Math.max(0, Math.ceil((state.timeLimitMs - state.elapsedMs) / 1000)),
    outcome: state.outcome,
    isPaused,
    damageDealt,
    damageTaken,
  };
}
