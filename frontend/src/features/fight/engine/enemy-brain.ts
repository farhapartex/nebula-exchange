import type { EnemyBrainProfile, FighterStats } from "@/features/fight/api/fight-setup-api";
import { attackPhaseOf } from "@/features/fight/engine/fight-simulation";
import type { AttackKind, FighterIntent, FighterRuntime } from "@/features/fight/engine/fight-types";
import { idleIntent } from "@/features/fight/engine/fight-types";

export type RandomNumber = () => number;

export type EnemyBrain = {
  currentIntent: FighterIntent;
  millisecondsUntilNextDecision: number;
};

export type BrainContext = {
  self: FighterRuntime;
  selfStats: FighterStats;
  opponent: FighterRuntime;
  opponentStats: FighterStats;
  profile: EnemyBrainProfile;
  random: RandomNumber;
};

export function createEnemyBrain(): EnemyBrain {
  return { currentIntent: idleIntent, millisecondsUntilNextDecision: 600 };
}

function pickAttack(profile: EnemyBrainProfile, random: RandomNumber): AttackKind {
  const totalWeight = profile.punch_weight + profile.kick_weight;
  return random() * totalWeight < profile.punch_weight ? "punch" : "kick";
}

export function decideEnemyIntent({
  self,
  selfStats,
  opponent,
  opponentStats,
  profile,
  random,
}: BrainContext): FighterIntent {
  const distance = Math.abs(opponent.x - self.x);
  const towardOpponent = opponent.x >= self.x ? 1 : -1;
  const opponentPhase = attackPhaseOf(opponent, opponentStats);
  const opponentThreatRange = opponent.attack ? opponentStats[opponent.attack].range + 12 : 0;

  if (opponentPhase === "windup" && distance <= opponentThreatRange) {
    const defensiveRoll = random();
    if (defensiveRoll < profile.dodge_chance && self.stamina >= selfStats.dodge_stamina_cost) {
      return { move: 0, attack: null, block: false, dodge: true };
    }
    if (defensiveRoll < profile.dodge_chance + profile.block_chance) {
      return { move: 0, attack: null, block: true, dodge: false };
    }
  }

  const attack = pickAttack(profile, random);
  const isInReach = distance <= selfStats[attack].range;
  if (isInReach && self.stamina >= selfStats[attack].stamina_cost && random() < profile.aggression) {
    return { move: 0, attack, block: false, dodge: false };
  }

  if (distance > profile.preferred_range + 18) {
    return { move: towardOpponent, attack: null, block: false, dodge: false };
  }
  if (distance < profile.preferred_range - 30) {
    return { move: -towardOpponent as -1 | 1, attack: null, block: false, dodge: false };
  }
  return random() < profile.block_chance * 0.5 ? { move: 0, attack: null, block: true, dodge: false } : idleIntent;
}

export function thinkEnemy(brain: EnemyBrain, context: BrainContext, deltaMs: number): FighterIntent {
  brain.millisecondsUntilNextDecision -= deltaMs;
  if (brain.millisecondsUntilNextDecision <= 0) {
    brain.currentIntent = decideEnemyIntent(context);
    brain.millisecondsUntilNextDecision = context.profile.reaction_ms * (0.6 + context.random() * 0.8);
  }
  const intent = brain.currentIntent;
  if (intent.attack || intent.dodge) {
    brain.currentIntent = { ...intent, attack: null, dodge: false };
  }
  return intent;
}
