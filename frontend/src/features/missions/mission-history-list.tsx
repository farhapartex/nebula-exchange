"use client";

import { useInfiniteQuery } from "@tanstack/react-query";
import { Rocket } from "lucide-react";

import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { StatusBadge } from "@/components/ui/status-badge";
import { ItemChip } from "@/features/catalog/components/item-chip";
import { useCatalogItems } from "@/features/catalog/use-catalog";
import { listMissions, missionsQueryKeys } from "@/features/missions/api/missions-api";
import { formatRelativeTime } from "@/utils/time/relative-time";

const historyPageSize = 10;

export function MissionHistoryList() {
  const { itemsByID } = useCatalogItems();
  const historyQuery = useInfiniteQuery({
    queryKey: missionsQueryKeys.history,
    queryFn: ({ pageParam }) => listMissions(["COLLECTED", "ABORTED"], { cursor: pageParam, limit: historyPageSize }),
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => lastPage.pagination.next_cursor,
  });
  const pastMissions = historyQuery.data?.pages.flatMap((missionPage) => missionPage.data) ?? [];

  if (historyQuery.isPending) {
    return <Skeleton className="h-40 rounded-2xl" />;
  }
  if (historyQuery.isError) {
    return <ErrorState message="We couldn't load your past missions." onRetry={() => void historyQuery.refetch()} />;
  }
  if (pastMissions.length === 0) {
    return (
      <EmptyState
        icon={Rocket}
        title="No finished missions yet"
        description="Collected and aborted runs show up here."
      />
    );
  }

  return (
    <div className="space-y-3">
      <ul className="divide-y divide-border rounded-2xl border border-border bg-surface/80">
        {pastMissions.map((pastMission) => (
          <li key={pastMission.id} className="flex flex-wrap items-center gap-3 px-4 py-3">
            <div className="min-w-40">
              <p className="text-sm font-medium text-foreground">{pastMission.zone_name}</p>
              <p className="text-xs text-muted">
                {formatRelativeTime(pastMission.collected_at ?? pastMission.aborted_at ?? pastMission.started_at)}
              </p>
            </div>
            {pastMission.status === "ABORTED" ? (
              <StatusBadge label="Aborted" tone="neutral" />
            ) : (
              <div className="flex flex-wrap gap-2">
                {pastMission.loot.map((roll) => {
                  const lootItem = itemsByID.get(roll.item_id);
                  return lootItem ? (
                    <ItemChip key={roll.item_id} item={lootItem} quantityLabel={String(roll.quantity)} />
                  ) : null;
                })}
              </div>
            )}
            <span className="ml-auto text-xs text-subtle">{pastMission.fuel_spent} fuel</span>
          </li>
        ))}
      </ul>
      {historyQuery.hasNextPage && (
        <Button
          variant="ghost"
          size="sm"
          isLoading={historyQuery.isFetchingNextPage}
          onClick={() => void historyQuery.fetchNextPage()}
        >
          Show more
        </Button>
      )}
    </div>
  );
}
