export type AttackKind = "punch" | "kick";

export type FighterAction = "idle" | "walking" | "attacking" | "blocking" | "dodging" | "stunned" | "knocked_out";

export type AttackPhase = "windup" | "active" | "recovery";

export type Facing = 1 | -1;

export type FighterIntent = {
  move: -1 | 0 | 1;
  attack: AttackKind | null;
  block: boolean;
  dodge: boolean;
};

export type FighterRuntime = {
  x: number;
  facing: Facing;
  health: number;
  stamina: number;
  action: FighterAction;
  actionElapsedMs: number;
  attack: AttackKind | null;
  hasLandedCurrentAttack: boolean;
  stunRemainingMs: number;
  dodgeDirection: Facing;
};

export type FightOutcome = "won" | "lost";

export type FightState = {
  player: FighterRuntime;
  enemy: FighterRuntime;
  elapsedMs: number;
  timeLimitMs: number;
  outcome: FightOutcome | null;
  leftBound: number;
  rightBound: number;
};

export type FighterRole = "player" | "enemy";

export type FightEvent =
  | {
      kind: "hit";
      attacker: FighterRole;
      defender: FighterRole;
      attack: AttackKind;
      damage: number;
      wasBlocked: boolean;
    }
  | { kind: "dodged"; attacker: FighterRole; defender: FighterRole }
  | { kind: "whiff"; attacker: FighterRole; attack: AttackKind }
  | { kind: "attack_started"; attacker: FighterRole; attack: AttackKind }
  | { kind: "knockout"; loser: FighterRole }
  | { kind: "time_up" };

export const idleIntent: FighterIntent = { move: 0, attack: null, block: false, dodge: false };
