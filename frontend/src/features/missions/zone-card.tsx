import { Check, Clock, Fuel, Lock, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { StatusBadge } from "@/components/ui/status-badge";
import type { CatalogItem, Zone } from "@/features/catalog/api/catalog-types";
import { LootRangeList } from "@/features/missions/loot-range-list";
import { describeZoneRequirements, expectedLootRanges } from "@/features/missions/mission-rules";
import { cn } from "@/utils/class-names";
import { formatDurationShort } from "@/utils/time/format-duration";

type ZoneCardProps = {
  zone: Zone;
  itemsByID: Map<number, CatalogItem>;
  isAtMissionLimit: boolean;
  onLaunch: (zone: Zone) => void;
};

function RequirementCheck({ label, isMet }: { label: string; isMet: boolean }) {
  const CheckIcon = isMet ? Check : X;
  return (
    <li className={cn("flex items-center gap-1.5 text-xs", isMet ? "text-up" : "text-muted")}>
      <CheckIcon className="size-3.5" aria-hidden="true" />
      {label}
    </li>
  );
}

export function ZoneCard({ zone, itemsByID, isAtMissionLimit, onLaunch }: ZoneCardProps) {
  const requirements = describeZoneRequirements(zone, itemsByID);
  const unlock = zone.unlock;
  const isLocked = unlock ? !unlock.is_unlocked : false;

  return (
    <article
      className={cn(
        "flex flex-col gap-4 rounded-2xl border bg-surface/80 p-5",
        isLocked ? "border-border" : "border-border-strong",
      )}
    >
      <div className="flex items-start justify-between gap-3">
        <div>
          <h3 className="text-base font-semibold text-foreground">{zone.name}</h3>
          <p className="mt-0.5 text-sm text-muted">{zone.description}</p>
        </div>
        {isLocked && <StatusBadge label="Locked" tone="neutral" />}
      </div>

      <dl className="flex flex-wrap gap-x-5 gap-y-1 text-sm">
        <div className="flex items-center gap-1.5 text-muted">
          <Clock className="size-4" aria-hidden="true" />
          <dt className="sr-only">Duration</dt>
          <dd className="text-foreground">{formatDurationShort(zone.duration_seconds)}</dd>
        </div>
        <div className="flex items-center gap-1.5 text-muted">
          <Fuel className="size-4" aria-hidden="true" />
          <dt className="sr-only">Fuel</dt>
          <dd className="text-foreground">{zone.fuel_cost} fuel</dd>
        </div>
        <div className="text-muted">
          <dt className="sr-only">Requirements</dt>
          <dd>{requirements.length > 0 ? requirements.join(" · ") : "No requirements"}</dd>
        </div>
      </dl>

      <div className="rounded-xl border border-border bg-background/50 p-3">
        <p className="mb-2 text-xs font-medium tracking-wide text-subtle uppercase">Base loot per run</p>
        <LootRangeList ranges={expectedLootRanges(zone.loot, null)} itemsByID={itemsByID} />
      </div>

      {unlock && isLocked && (
        <ul className="space-y-1">
          <RequirementCheck label={`A free drill, Tier ${zone.minimum_drill_tier} or better`} isMet={unlock.has_required_drill} />
          <RequirementCheck label="A free ship that can fly here" isMet={unlock.has_allowed_ship} />
          <RequirementCheck label={`${zone.fuel_cost} Fuel Cells`} isMet={unlock.has_enough_fuel} />
        </ul>
      )}

      <Button
        className="mt-auto"
        variant={isLocked ? "secondary" : "primary"}
        disabled={isLocked || isAtMissionLimit}
        onClick={() => onLaunch(zone)}
      >
        {isLocked && <Lock className="size-4" aria-hidden="true" />}
        {isAtMissionLimit && !isLocked ? "3 missions running" : "Launch mission"}
      </Button>
    </article>
  );
}
