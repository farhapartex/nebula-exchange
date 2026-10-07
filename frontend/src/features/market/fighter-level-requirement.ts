export function levelsStillNeeded(minimumFighterLevel: number, currentFighterLevel: number | undefined): number {
  if (currentFighterLevel === undefined) {
    return 0;
  }
  return Math.max(minimumFighterLevel - currentFighterLevel, 0);
}

export function describeLevelsStillNeeded(levelsNeeded: number): string {
  return levelsNeeded === 1 ? "Earn 1 more level" : `Earn ${levelsNeeded} more levels`;
}
