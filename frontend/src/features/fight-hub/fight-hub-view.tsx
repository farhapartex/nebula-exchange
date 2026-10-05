import { FighterCard } from "@/features/fight-hub/fighter-card";
import { LoadoutCard } from "@/features/fight-hub/loadout-card";
import { LockedChapterCard } from "@/features/fight-hub/locked-chapter-card";
import { NextFightCard } from "@/features/fight-hub/next-fight-card";
import { TrainingCard } from "@/features/fight-hub/training-card";

export function FightHubView() {
  return (
    <div className="grid gap-5 lg:grid-cols-3">
      <div className="lg:col-span-1">
        <FighterCard />
      </div>
      <div className="lg:col-span-2">
        <NextFightCard />
      </div>
      <div className="lg:col-span-2">
        <LoadoutCard />
      </div>
      <div className="space-y-5">
        <LockedChapterCard />
        <TrainingCard />
      </div>
    </div>
  );
}
