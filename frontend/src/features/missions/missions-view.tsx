"use client";

import { useState } from "react";

import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import type { Zone } from "@/features/catalog/api/catalog-types";
import { useCatalogItems, useZones } from "@/features/catalog/use-catalog";
import { useInventory } from "@/features/inventory/use-inventory";
import { ActiveMissions, useActiveMissions } from "@/features/missions/active-missions";
import { MissionHistoryList } from "@/features/missions/mission-history-list";
import { MissionLauncher } from "@/features/missions/mission-launcher";
import { maximumActiveMissions } from "@/features/missions/mission-rules";
import { ZoneCard } from "@/features/missions/zone-card";

export function MissionsView() {
  const zonesQuery = useZones();
  const { itemsByID } = useCatalogItems();
  const inventoryQuery = useInventory();
  const activeMissionsQuery = useActiveMissions();
  const [zoneBeingLaunched, setZoneBeingLaunched] = useState<Zone | null>(null);
  const activeMissions = activeMissionsQuery.data ?? [];

  return (
    <div className="space-y-10">
      <section>
        <div className="mb-3 flex items-baseline justify-between gap-3">
          <h2 className="text-lg font-semibold text-foreground">Active missions</h2>
          <span className="text-sm text-muted tabular-nums">
            {activeMissions.length} / {maximumActiveMissions}
          </span>
        </div>
        {activeMissionsQuery.isPending ? (
          <Skeleton className="h-24 rounded-2xl" />
        ) : (
          <ActiveMissions missions={activeMissions} itemsByID={itemsByID} />
        )}
      </section>

      <section>
        <h2 className="mb-3 text-lg font-semibold text-foreground">Zones</h2>
        {zonesQuery.isPending && (
          <div className="grid gap-4 lg:grid-cols-2">
            {Array.from({ length: 4 }, (_, skeletonIndex) => (
              <Skeleton key={skeletonIndex} className="h-72 rounded-2xl" />
            ))}
          </div>
        )}
        {zonesQuery.isError && (
          <ErrorState message="We couldn't load the zones." onRetry={() => void zonesQuery.refetch()} />
        )}
        {zonesQuery.data && (
          <div className="grid gap-4 lg:grid-cols-2">
            {zonesQuery.data.map((zone) => (
              <ZoneCard
                key={zone.id}
                zone={zone}
                itemsByID={itemsByID}
                isAtMissionLimit={activeMissions.length >= maximumActiveMissions}
                onLaunch={setZoneBeingLaunched}
              />
            ))}
          </div>
        )}
      </section>

      <section>
        <h2 className="mb-3 text-lg font-semibold text-foreground">Mission history</h2>
        <MissionHistoryList />
      </section>

      <MissionLauncher
        zone={zoneBeingLaunched}
        holdings={inventoryQuery.data ?? []}
        itemsByID={itemsByID}
        onClose={() => setZoneBeingLaunched(null)}
      />
    </div>
  );
}
