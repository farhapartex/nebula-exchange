import type { AttackKind, FighterIntent, FightOutcome } from "@/features/fight/engine/fight-types";

export type FighterHudState = {
  health: number;
  maxHealth: number;
  stamina: number;
  maxStamina: number;
};

export type FightSnapshot = {
  player: FighterHudState;
  enemy: FighterHudState;
  remainingSeconds: number;
  elapsedMs: number;
  outcome: FightOutcome | null;
  isPaused: boolean;
  damageDealt: number;
  damageTaken: number;
};

export type HeldControl = "left" | "right" | "block";
export type TappedControl = AttackKind | "dodge";

type SnapshotListener = () => void;

export class FightController {
  private heldControls = new Set<HeldControl>();
  private pendingTap: TappedControl | null = null;
  private snapshot: FightSnapshot;
  private listeners = new Set<SnapshotListener>();
  private restartHandler: (() => void) | null = null;

  constructor(initialSnapshot: FightSnapshot) {
    this.snapshot = initialSnapshot;
  }

  press(control: HeldControl): void {
    this.heldControls.add(control);
  }

  release(control: HeldControl): void {
    this.heldControls.delete(control);
  }

  tap(control: TappedControl): void {
    this.pendingTap = control;
  }

  releaseEverything(): void {
    this.heldControls.clear();
    this.pendingTap = null;
  }

  readPlayerIntent(): FighterIntent {
    const isLeft = this.heldControls.has("left");
    const isRight = this.heldControls.has("right");
    const tap = this.pendingTap;
    this.pendingTap = null;
    return {
      move: isLeft === isRight ? 0 : isLeft ? -1 : 1,
      attack: tap === "punch" || tap === "kick" ? tap : null,
      block: this.heldControls.has("block"),
      dodge: tap === "dodge",
    };
  }

  togglePause(): void {
    if (this.snapshot.outcome) {
      return;
    }
    this.publish({ ...this.snapshot, isPaused: !this.snapshot.isPaused });
    this.releaseEverything();
  }

  get isPaused(): boolean {
    return this.snapshot.isPaused;
  }

  onRestart(restartHandler: (() => void) | null): void {
    this.restartHandler = restartHandler;
  }

  restart(): void {
    this.releaseEverything();
    this.snapshot = { ...this.snapshot, isPaused: false, outcome: null };
    this.restartHandler?.();
  }

  publish(nextSnapshot: FightSnapshot): void {
    this.snapshot = nextSnapshot;
    for (const listener of this.listeners) {
      listener();
    }
  }

  getSnapshot = (): FightSnapshot => this.snapshot;

  subscribe = (listener: SnapshotListener): (() => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };
}
