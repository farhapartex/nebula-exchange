"use client";

import { useInfiniteQuery } from "@tanstack/react-query";
import { Wrench } from "lucide-react";

import { NcAmount } from "@/components/money/nc-amount";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { StatusBadge } from "@/components/ui/status-badge";
import { ItemChip } from "@/features/catalog/components/item-chip";
import { useCatalogItems } from "@/features/catalog/use-catalog";
import { listCrafts, workshopQueryKeys } from "@/features/workshop/api/workshop-api";
import { formatRelativeTime } from "@/utils/time/relative-time";

export function CraftHistoryList() {
  const { itemsByID } = useCatalogItems();
  const craftsQuery = useInfiniteQuery({
    queryKey: workshopQueryKeys.craftHistory,
    queryFn: ({ pageParam }) => listCrafts(null, { cursor: pageParam, limit: 20 }),
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => lastPage.pagination.next_cursor,
  });
  const craftJobs = craftsQuery.data?.pages.flatMap((craftPage) => craftPage.data) ?? [];

  if (craftsQuery.isPending) {
    return <Skeleton className="h-40 rounded-2xl" />;
  }
  if (craftsQuery.isError) {
    return <ErrorState message="We couldn't load your crafts." onRetry={() => void craftsQuery.refetch()} />;
  }
  if (craftJobs.length === 0) {
    return <EmptyState icon={Wrench} title="Nothing crafted yet" description="Your crafting jobs show up here." />;
  }
  return (
    <div className="space-y-3">
      <ul className="divide-y divide-border rounded-2xl border border-border bg-surface/80">
        {craftJobs.map((craftJob) => {
          const outputItem = itemsByID.get(craftJob.output_item_id);
          return (
            <li key={craftJob.id} className="flex flex-wrap items-center gap-3 px-4 py-3">
              {outputItem && <ItemChip item={outputItem} quantityLabel={String(craftJob.output_quantity)} />}
              <StatusBadge
                label={craftJob.status === "DELIVERED" ? "Delivered" : "Crafting"}
                tone={craftJob.status === "DELIVERED" ? "success" : "info"}
              />
              <span className="text-xs text-muted">
                {formatRelativeTime(craftJob.delivered_at ?? craftJob.started_at)}
              </span>
              <span className="ml-auto text-xs text-muted">
                fee <NcAmount amount={craftJob.fee} />
              </span>
            </li>
          );
        })}
      </ul>
      {craftsQuery.hasNextPage && (
        <Button
          variant="ghost"
          size="sm"
          isLoading={craftsQuery.isFetchingNextPage}
          onClick={() => void craftsQuery.fetchNextPage()}
        >
          Show more
        </Button>
      )}
    </div>
  );
}
