"use client";

import { Loader2 } from "lucide-react";

import { CountdownTimer } from "@/components/time/countdown-timer";
import { Button } from "@/components/ui/button";
import { StatusBadge } from "@/components/ui/status-badge";
import type { CatalogItem } from "@/features/catalog/api/catalog-types";
import { ItemIcon } from "@/features/catalog/components/item-icon";
import type { Mission } from "@/features/missions/api/missions-api";
import { useNow } from "@/utils/time/use-now";

type ActiveMissionCardProps = {
  mission: Mission;
  itemsByID: Map<number, CatalogItem>;
  isCollecting: boolean;
  onCollect: (mission: Mission) => void;
  onAbort: (mission: Mission) => void;
};

export function ActiveMissionCard({ mission, itemsByID, isCollecting, onCollect, onAbort }: ActiveMissionCardProps) {
  const now = useNow();
  const startedAt = new Date(mission.started_at).getTime();
  const endsAt = new Date(mission.ends_at).getTime();
  const progress = Math.min(Math.max((now - startedAt) / (endsAt - startedAt), 0), 1);
  const hasArrived = now >= endsAt;
  const ship = itemsByID.get(mission.ship_item_id);
  const drill = itemsByID.get(mission.drill_item_id);

  return (
    <li className="rounded-2xl border border-border bg-surface/80 p-4">
      <div className="flex flex-wrap items-center gap-3">
        <div className="flex -space-x-2">
          {ship && <ItemIcon item={ship} size="sm" className="ring-2 ring-surface" />}
          {drill && <ItemIcon item={drill} size="sm" className="ring-2 ring-surface" />}
        </div>
        <div className="min-w-0 flex-1">
          <p className="text-sm font-semibold text-foreground">{mission.zone_name}</p>
          <p className="text-xs text-muted">
            {ship?.name} · {drill?.name} · {mission.fuel_spent} fuel
          </p>
        </div>
        {mission.status === "COMPLETED" ? (
          <Button size="sm" isLoading={isCollecting} onClick={() => onCollect(mission)}>
            Collect
          </Button>
        ) : hasArrived ? (
          <StatusBadge label="Unloading…" tone="info" isPulsing />
        ) : (
          <>
            <CountdownTimer endsAt={mission.ends_at} className="text-sm text-foreground" />
            <Button size="sm" variant="ghost" onClick={() => onAbort(mission)}>
              Abort
            </Button>
          </>
        )}
      </div>
      <div
        className="mt-3 h-1.5 overflow-hidden rounded-full bg-border"
        role="progressbar"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={Math.round(progress * 100)}
      >
        <div
          className={
            mission.status === "COMPLETED" ? "h-full bg-up" : "h-full bg-accent transition-[width] duration-1000"
          }
          style={{ width: `${progress * 100}%` }}
        />
      </div>
      {hasArrived && mission.status === "RUNNING" && (
        <p className="mt-2 flex items-center gap-1.5 text-xs text-muted">
          <Loader2 className="size-3 animate-spin" aria-hidden="true" /> Counting the haul, a few seconds…
        </p>
      )}
    </li>
  );
}
