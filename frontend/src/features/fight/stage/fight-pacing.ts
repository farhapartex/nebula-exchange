export type FightPacing = {
  hitStopRemainingMs: number;
  slowMotionRemainingMs: number;
};

const slowMotionTimeScale = 0.25;

export function createFightPacing(): FightPacing {
  return { hitStopRemainingMs: 0, slowMotionRemainingMs: 0 };
}

export function freezeForImpact(pacing: FightPacing, durationMs: number): void {
  pacing.hitStopRemainingMs = Math.max(pacing.hitStopRemainingMs, durationMs);
}

export function startSlowMotion(pacing: FightPacing, durationMs: number): void {
  pacing.slowMotionRemainingMs = Math.max(pacing.slowMotionRemainingMs, durationMs);
}

export function isInSlowMotion(pacing: FightPacing): boolean {
  return pacing.slowMotionRemainingMs > 0;
}

export function advancePacing(pacing: FightPacing, realDeltaMs: number): number {
  if (pacing.hitStopRemainingMs > 0) {
    pacing.hitStopRemainingMs = Math.max(0, pacing.hitStopRemainingMs - realDeltaMs);
    return 0;
  }
  if (pacing.slowMotionRemainingMs > 0) {
    pacing.slowMotionRemainingMs = Math.max(0, pacing.slowMotionRemainingMs - realDeltaMs);
    return realDeltaMs * slowMotionTimeScale;
  }
  return realDeltaMs;
}
