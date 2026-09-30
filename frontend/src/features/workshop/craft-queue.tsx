"use client";

import { useEffect, useRef } from "react";
import { useQuery } from "@tanstack/react-query";
import { Loader2, Wrench } from "lucide-react";

import { CountdownTimer } from "@/components/time/countdown-timer";
import { useToast } from "@/components/ui/toast/use-toast";
import type { CatalogItem } from "@/features/catalog/api/catalog-types";
import { ItemIcon } from "@/features/catalog/components/item-icon";
import { listCrafts, workshopQueryKeys, type CraftJob } from "@/features/workshop/api/workshop-api";
import { useRefreshAfterWorkshopChange } from "@/features/workshop/use-refresh-after-workshop-change";
import { useNow } from "@/utils/time/use-now";

const craftRefreshIntervalInMilliseconds = 5_000;

export function useActiveCraft() {
  return useQuery({
    queryKey: workshopQueryKeys.activeCraft,
    queryFn: async () => (await listCrafts("CRAFTING", { limit: 1 })).data[0] ?? null,
    refetchInterval: (query) => (query.state.data ? craftRefreshIntervalInMilliseconds : false),
  });
}

function useDeliveryToast(activeCraft: CraftJob | null | undefined, itemsByID: Map<number, CatalogItem>) {
  const { showToast } = useToast();
  const refreshAfterWorkshopChange = useRefreshAfterWorkshopChange();
  const previousCraft = useRef<CraftJob | null | undefined>(undefined);

  useEffect(() => {
    const finishedCraft = previousCraft.current;
    previousCraft.current = activeCraft;
    if (finishedCraft && activeCraft === null) {
      refreshAfterWorkshopChange();
      showToast({
        tone: "success",
        title: `${finishedCraft.output_quantity} × ${itemsByID.get(finishedCraft.output_item_id)?.name ?? "item"} ready`,
        description: "Your craft is in your inventory.",
      });
    }
  }, [activeCraft, itemsByID, refreshAfterWorkshopChange, showToast]);
}

export function CraftQueue({ itemsByID }: { itemsByID: Map<number, CatalogItem> }) {
  const activeCraftQuery = useActiveCraft();
  const activeCraft = activeCraftQuery.data;
  const now = useNow();
  useDeliveryToast(activeCraft, itemsByID);

  if (!activeCraft) {
    return (
      <div className="flex items-center gap-3 rounded-2xl border border-dashed border-border-strong px-4 py-4 text-sm text-muted">
        <Wrench className="size-4 text-subtle" aria-hidden="true" />
        The workshop is free. Pick a recipe to start crafting.
      </div>
    );
  }

  const outputItem = itemsByID.get(activeCraft.output_item_id);
  const startedAt = new Date(activeCraft.started_at).getTime();
  const endsAt = new Date(activeCraft.ends_at).getTime();
  const progress = Math.min(Math.max((now - startedAt) / (endsAt - startedAt), 0), 1);

  return (
    <div className="rounded-2xl border border-accent/40 bg-accent/5 p-4">
      <div className="flex flex-wrap items-center gap-3">
        {outputItem && <ItemIcon item={outputItem} />}
        <div className="min-w-0 flex-1">
          <p className="text-sm font-semibold text-foreground">
            Crafting {activeCraft.output_quantity} × {outputItem?.name}
          </p>
          <p className="text-xs text-muted">One craft at a time. The next one can start when this finishes.</p>
        </div>
        {now < endsAt ? (
          <CountdownTimer endsAt={activeCraft.ends_at} className="text-lg text-foreground" />
        ) : (
          <span className="flex items-center gap-1.5 text-sm text-muted">
            <Loader2 className="size-4 animate-spin" aria-hidden="true" /> Delivering…
          </span>
        )}
      </div>
      <div
        className="mt-3 h-1.5 overflow-hidden rounded-full bg-border"
        role="progressbar"
        aria-valuenow={Math.round(progress * 100)}
        aria-valuemin={0}
        aria-valuemax={100}
      >
        <div className="h-full bg-accent transition-[width] duration-1000" style={{ width: `${progress * 100}%` }} />
      </div>
    </div>
  );
}
