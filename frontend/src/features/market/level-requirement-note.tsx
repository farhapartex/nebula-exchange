import { TrendingUp } from "lucide-react";

import { describeLevelsStillNeeded } from "@/features/market/fighter-level-requirement";

type LevelRequirementNoteProps = {
  minimumFighterLevel: number;
  levelsNeeded: number;
};

export function LevelRequirementNote({ minimumFighterLevel, levelsNeeded }: LevelRequirementNoteProps) {
  return (
    <span className="flex flex-col items-end text-right">
      <span className="flex items-center gap-1.5 text-xs font-medium text-amber-300">
        <TrendingUp className="size-3.5" aria-hidden="true" />
        {describeLevelsStillNeeded(levelsNeeded)}
      </span>
      <span className="text-[0.6875rem] text-subtle">Unlocks at fighter level {minimumFighterLevel}</span>
    </span>
  );
}
