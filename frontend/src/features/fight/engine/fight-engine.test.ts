import { describe, expect, it } from "vitest";

import type { EnemyBrainProfile, FighterStats } from "@/features/fight/api/fight-setup-api";
import { decideEnemyIntent } from "@/features/fight/engine/enemy-brain";
import { createFightState, stepFight, type FightRules } from "@/features/fight/engine/fight-simulation";
import type { FighterIntent, FightState } from "@/features/fight/engine/fight-types";
import { idleIntent } from "@/features/fight/engine/fight-types";

const baseStats: FighterStats = {
  max_health: 100,
  max_stamina: 100,
  stamina_regen_per_second: 20,
  walk_speed: 200,
  block_damage_reduction: 0.75,
  block_stamina_cost: 10,
  dodge_stamina_cost: 20,
  dodge_duration_ms: 300,
  dodge_distance: 120,
  punch: {
    damage: 10,
    stamina_cost: 8,
    range: 80,
    windup_ms: 100,
    active_ms: 80,
    recovery_ms: 150,
    knockback: 10,
    hit_stun_ms: 250,
  },
  kick: {
    damage: 16,
    stamina_cost: 14,
    range: 110,
    windup_ms: 220,
    active_ms: 100,
    recovery_ms: 260,
    knockback: 24,
    hit_stun_ms: 380,
  },
};

const rules: FightRules = { player: baseStats, enemy: baseStats };

function closeQuarters(): FightState {
  const state = createFightState(rules, 960, 60);
  state.player.x = 400;
  state.enemy.x = 460;
  return state;
}

function run(state: FightState, playerIntent: FighterIntent, enemyIntent: FighterIntent, totalMs: number) {
  const events = [];
  for (let elapsed = 0; elapsed < totalMs; elapsed += 16) {
    events.push(...stepFight(state, rules, { player: playerIntent, enemy: enemyIntent }, 16));
    playerIntent = { ...playerIntent, attack: null, dodge: false };
    enemyIntent = { ...enemyIntent, attack: null, dodge: false };
  }
  return events;
}

describe("fight simulation", () => {
  it("lands a punch in range, damages and stuns the defender", () => {
    const state = closeQuarters();
    const events = run(state, { ...idleIntent, attack: "punch" }, idleIntent, 200);
    expect(events).toContainEqual(
      expect.objectContaining({ kind: "hit", attacker: "player", damage: 10, wasBlocked: false }),
    );
    expect(state.enemy.health).toBe(90);
    expect(state.enemy.action).toBe("stunned");
    expect(state.player.stamina).toBeLessThan(100);
  });

  it("reduces damage when the defender blocks facing the attack", () => {
    const state = closeQuarters();
    run(state, { ...idleIntent, attack: "punch" }, { ...idleIntent, block: true }, 200);
    expect(state.enemy.health).toBe(97);
    expect(state.enemy.action).toBe("blocking");
  });

  it("makes a dodging fighter untouchable", () => {
    const state = closeQuarters();
    const events = run(state, { ...idleIntent, attack: "kick" }, { ...idleIntent, dodge: true }, 700);
    expect(state.enemy.health).toBe(100);
    expect(events.some((event) => event.kind === "dodged" || event.kind === "whiff")).toBe(true);
  });

  it("misses when out of range and never lets fighters overlap or leave the arena", () => {
    const state = createFightState(rules, 960, 60);
    const events = run(state, { ...idleIntent, attack: "punch" }, idleIntent, 400);
    expect(events).toContainEqual({ kind: "whiff", attacker: "player", attack: "punch" });
    run(state, { ...idleIntent, move: 1 }, { ...idleIntent, move: -1 }, 3000);
    expect(state.enemy.x - state.player.x).toBeGreaterThanOrEqual(45);
    run(state, { ...idleIntent, move: -1 }, idleIntent, 6000);
    expect(state.player.x).toBeGreaterThanOrEqual(state.leftBound);
  });

  it("ends in a knockout and ignores input afterwards", () => {
    const state = closeQuarters();
    state.enemy.health = 5;
    const events = run(state, { ...idleIntent, attack: "punch" }, idleIntent, 200);
    expect(events).toContainEqual({ kind: "knockout", loser: "enemy" });
    expect(state.outcome).toBe("won");
    expect(stepFight(state, rules, { player: { ...idleIntent, attack: "punch" }, enemy: idleIntent }, 16)).toEqual([]);
  });

  it("decides on health share when time runs out", () => {
    const state = closeQuarters();
    state.timeLimitMs = 100;
    state.player.health = 40;
    state.enemy.health = 60;
    run(state, idleIntent, idleIntent, 200);
    expect(state.outcome).toBe("lost");
  });

  it("refuses attacks the fighter has no stamina for", () => {
    const state = closeQuarters();
    state.player.stamina = 2;
    run(state, { ...idleIntent, attack: "punch" }, idleIntent, 200);
    expect(state.enemy.health).toBe(100);
  });
});

describe("enemy brain", () => {
  const profile: EnemyBrainProfile = {
    aggression: 1,
    block_chance: 1,
    dodge_chance: 0,
    preferred_range: 80,
    reaction_ms: 300,
    punch_weight: 1,
    kick_weight: 0,
  };

  it("walks toward a distant opponent and attacks when in reach", () => {
    const state = createFightState(rules, 960, 60);
    const context = {
      self: state.enemy,
      selfStats: baseStats,
      opponent: state.player,
      opponentStats: baseStats,
      profile,
      random: () => 0,
    };
    expect(decideEnemyIntent(context).move).toBe(-1);
    state.enemy.x = state.player.x + 60;
    expect(decideEnemyIntent(context).attack).toBe("punch");
  });

  it("blocks an incoming wind-up when it is cautious", () => {
    const state = closeQuarters();
    state.player.action = "attacking";
    state.player.attack = "kick";
    state.player.actionElapsedMs = 20;
    const context = {
      self: state.enemy,
      selfStats: baseStats,
      opponent: state.player,
      opponentStats: baseStats,
      profile: { ...profile, aggression: 0 },
      random: () => 0.1,
    };
    expect(decideEnemyIntent(context).block).toBe(true);
  });
});
