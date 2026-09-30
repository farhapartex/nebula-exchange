"use client";

import { Coins, Gift } from "lucide-react";

import { NcAmount } from "@/components/money/nc-amount";
import { Skeleton } from "@/components/ui/skeleton";
import { ItemIcon } from "@/features/catalog/components/item-icon";
import { useCatalogItems } from "@/features/catalog/use-catalog";
import { starterBonusInMicroUnits, starterPackItems } from "@/features/onboarding/starter-pack";

export function StarterPackPreview() {
  const { itemsByID, isPending } = useCatalogItems();

  return (
    <div className="rounded-2xl border border-border bg-surface/70 p-5">
      <p className="mb-4 flex items-center gap-2 text-sm font-medium text-foreground">
        <Gift className="size-4 text-accent-soft" aria-hidden="true" />
        Your starter pack
      </p>
      <ul className="grid gap-3 sm:grid-cols-2">
        {starterPackItems.map((starterPackItem) => {
          const catalogItem = itemsByID.get(starterPackItem.itemID);
          return (
            <li
              key={starterPackItem.itemID}
              className="flex items-center gap-3 rounded-xl border border-border bg-background/60 p-3"
            >
              {catalogItem ? <ItemIcon item={catalogItem} /> : <Skeleton className="size-12 rounded-xl" />}
              <div className="min-w-0">
                <p className="text-sm text-foreground">
                  <span className="font-mono text-accent-soft">{starterPackItem.quantity}×</span>{" "}
                  {catalogItem?.name ?? (isPending ? "Loading…" : "Item")}
                </p>
                <p className="mt-0.5 text-xs text-muted">{starterPackItem.description}</p>
              </div>
            </li>
          );
        })}
        <li className="flex items-center gap-3 rounded-xl border border-border bg-background/60 p-3">
          <span className="flex size-12 shrink-0 items-center justify-center rounded-xl border border-white/10 bg-linear-to-br from-amber-400/40 to-amber-900/60">
            <Coins className="size-6 text-amber-100" aria-hidden="true" />
          </span>
          <div className="min-w-0">
            <p className="text-sm text-foreground">
              <NcAmount amount={starterBonusInMicroUnits} className="text-accent-soft" /> bonus
            </p>
            <p className="mt-0.5 text-xs text-muted">Card NC to get you going</p>
          </div>
        </li>
      </ul>
    </div>
  );
}
