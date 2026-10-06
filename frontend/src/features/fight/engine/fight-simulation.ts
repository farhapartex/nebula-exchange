import type { AttackStats, FighterStats } from "@/features/fight/api/fight-setup-api";
import type {
  AttackKind,
  AttackPhase,
  Facing,
  FightEvent,
  FighterIntent,
  FighterRole,
  FighterRuntime,
  FightState,
} from "@/features/fight/engine/fight-types";

export const minimumFighterSeparation = 46;

export type FightRules = {
  player: FighterStats;
  enemy: FighterStats;
};

export function createFighter(stats: FighterStats, x: number, facing: Facing): FighterRuntime {
  return {
    x,
    facing,
    health: stats.max_health,
    stamina: stats.max_stamina,
    action: "idle",
    actionElapsedMs: 0,
    attack: null,
    hasLandedCurrentAttack: false,
    stunRemainingMs: 0,
    dodgeDirection: facing,
  };
}

export function createFightState(rules: FightRules, arenaWidth: number, timeLimitSeconds: number): FightState {
  return {
    player: createFighter(rules.player, arenaWidth * 0.32, 1),
    enemy: createFighter(rules.enemy, arenaWidth * 0.68, -1),
    elapsedMs: 0,
    timeLimitMs: timeLimitSeconds * 1000,
    outcome: null,
    leftBound: 40,
    rightBound: arenaWidth - 40,
  };
}

export function attackPhaseOf(fighter: FighterRuntime, stats: FighterStats): AttackPhase | null {
  if (fighter.action !== "attacking" || !fighter.attack) {
    return null;
  }
  const attackStats = stats[fighter.attack];
  if (fighter.actionElapsedMs < attackStats.windup_ms) {
    return "windup";
  }
  if (fighter.actionElapsedMs < attackStats.windup_ms + attackStats.active_ms) {
    return "active";
  }
  return "recovery";
}

function attackDuration(attackStats: AttackStats): number {
  return attackStats.windup_ms + attackStats.active_ms + attackStats.recovery_ms;
}

function isBusy(fighter: FighterRuntime): boolean {
  return (
    fighter.action === "attacking" ||
    fighter.action === "dodging" ||
    fighter.action === "stunned" ||
    fighter.action === "knocked_out"
  );
}

function startAction(fighter: FighterRuntime, action: FighterRuntime["action"]): void {
  fighter.action = action;
  fighter.actionElapsedMs = 0;
}

function applyIntent(
  fighter: FighterRuntime,
  stats: FighterStats,
  intent: FighterIntent,
  role: FighterRole,
  events: FightEvent[],
): void {
  if (isBusy(fighter)) {
    return;
  }
  if (intent.dodge && fighter.stamina >= stats.dodge_stamina_cost) {
    fighter.stamina -= stats.dodge_stamina_cost;
    fighter.dodgeDirection = intent.move === 0 ? (-fighter.facing as Facing) : intent.move;
    startAction(fighter, "dodging");
    return;
  }
  if (intent.attack && fighter.stamina >= stats[intent.attack].stamina_cost) {
    fighter.stamina -= stats[intent.attack].stamina_cost;
    fighter.attack = intent.attack;
    fighter.hasLandedCurrentAttack = false;
    startAction(fighter, "attacking");
    events.push({ kind: "attack_started", attacker: role, attack: intent.attack });
    return;
  }
  if (intent.block) {
    if (fighter.action !== "blocking") {
      startAction(fighter, "blocking");
    }
    return;
  }
  const nextAction = intent.move === 0 ? "idle" : "walking";
  if (fighter.action !== nextAction) {
    startAction(fighter, nextAction);
  }
}

function advanceFighter(
  fighter: FighterRuntime,
  stats: FighterStats,
  intent: FighterIntent,
  deltaMs: number,
  role: FighterRole,
  events: FightEvent[],
): void {
  fighter.actionElapsedMs += deltaMs;
  const deltaSeconds = deltaMs / 1000;

  switch (fighter.action) {
    case "walking":
      fighter.x += intent.move * stats.walk_speed * deltaSeconds;
      break;
    case "dodging":
      fighter.x += fighter.dodgeDirection * (stats.dodge_distance / stats.dodge_duration_ms) * deltaMs;
      if (fighter.actionElapsedMs >= stats.dodge_duration_ms) {
        startAction(fighter, "idle");
      }
      break;
    case "stunned":
      fighter.stunRemainingMs -= deltaMs;
      if (fighter.stunRemainingMs <= 0) {
        startAction(fighter, "idle");
      }
      break;
    case "attacking":
      if (fighter.attack && fighter.actionElapsedMs >= attackDuration(stats[fighter.attack])) {
        if (!fighter.hasLandedCurrentAttack) {
          events.push({ kind: "whiff", attacker: role, attack: fighter.attack });
        }
        fighter.attack = null;
        startAction(fighter, "idle");
      }
      break;
  }

  const isRecovering = fighter.action === "idle" || fighter.action === "walking";
  if (isRecovering) {
    fighter.stamina = Math.min(stats.max_stamina, fighter.stamina + stats.stamina_regen_per_second * deltaSeconds);
  }
}

function faceOpponent(fighter: FighterRuntime, opponent: FighterRuntime): void {
  if (fighter.action === "attacking" || fighter.action === "dodging" || fighter.action === "knocked_out") {
    return;
  }
  fighter.facing = opponent.x >= fighter.x ? 1 : -1;
}

function resolveAttack(
  attacker: FighterRuntime,
  attackerStats: FighterStats,
  defender: FighterRuntime,
  defenderStats: FighterStats,
  attackerRole: FighterRole,
  defenderRole: FighterRole,
  events: FightEvent[],
): void {
  if (attackPhaseOf(attacker, attackerStats) !== "active" || attacker.hasLandedCurrentAttack || !attacker.attack) {
    return;
  }
  const attackStats = attackerStats[attacker.attack];
  const distance = (defender.x - attacker.x) * attacker.facing;
  if (distance < 0 || distance > attackStats.range || defender.action === "knocked_out") {
    return;
  }
  attacker.hasLandedCurrentAttack = true;

  if (defender.action === "dodging") {
    events.push({ kind: "dodged", attacker: attackerRole, defender: defenderRole });
    return;
  }

  const isBlocked = defender.action === "blocking" && defender.facing === -attacker.facing;
  const damage = Math.round(
    isBlocked ? attackStats.damage * (1 - defenderStats.block_damage_reduction) : attackStats.damage,
  );
  defender.health = Math.max(0, defender.health - damage);
  defender.x += attacker.facing * attackStats.knockback * (isBlocked ? 0.35 : 1);
  if (isBlocked) {
    defender.stamina = Math.max(0, defender.stamina - defenderStats.block_stamina_cost);
  } else {
    defender.attack = null;
    defender.stunRemainingMs = attackStats.hit_stun_ms;
    startAction(defender, "stunned");
  }
  events.push({
    kind: "hit",
    attacker: attackerRole,
    defender: defenderRole,
    attack: attacker.attack as AttackKind,
    damage,
    wasBlocked: isBlocked,
  });
}

function keepApartAndInBounds(state: FightState): void {
  const { player, enemy } = state;
  for (const fighter of [player, enemy]) {
    fighter.x = Math.min(state.rightBound, Math.max(state.leftBound, fighter.x));
  }
  const separation = enemy.x - player.x;
  if (Math.abs(separation) < minimumFighterSeparation) {
    const push = (minimumFighterSeparation - Math.abs(separation)) / 2;
    const direction = separation >= 0 ? 1 : -1;
    player.x -= push * direction;
    enemy.x += push * direction;
  }
}

function decideOutcome(state: FightState, events: FightEvent[]): void {
  if (state.player.health <= 0 || state.enemy.health <= 0) {
    const loser: FighterRole = state.player.health <= 0 ? "player" : "enemy";
    startAction(state[loser], "knocked_out");
    state.outcome = loser === "enemy" ? "won" : "lost";
    events.push({ kind: "knockout", loser });
    return;
  }
  if (state.elapsedMs >= state.timeLimitMs) {
    state.outcome = "lost";
    events.push({ kind: "time_up" });
  }
}

export function stepFight(
  state: FightState,
  rules: FightRules,
  intents: { player: FighterIntent; enemy: FighterIntent },
  deltaMs: number,
): FightEvent[] {
  const events: FightEvent[] = [];
  if (state.outcome) {
    return events;
  }
  state.elapsedMs += deltaMs;

  applyIntent(state.player, rules.player, intents.player, "player", events);
  applyIntent(state.enemy, rules.enemy, intents.enemy, "enemy", events);
  advanceFighter(state.player, rules.player, intents.player, deltaMs, "player", events);
  advanceFighter(state.enemy, rules.enemy, intents.enemy, deltaMs, "enemy", events);
  faceOpponent(state.player, state.enemy);
  faceOpponent(state.enemy, state.player);
  resolveAttack(state.player, rules.player, state.enemy, rules.enemy, "player", "enemy", events);
  resolveAttack(state.enemy, rules.enemy, state.player, rules.player, "enemy", "player", events);
  keepApartAndInBounds(state);
  decideOutcome(state, events);
  return events;
}
