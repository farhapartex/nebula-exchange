"use client";

import { useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { Coins } from "lucide-react";
import { Tabs } from "radix-ui";

import { useFighterProfile } from "@/features/fight-hub/use-fight-hub";
import { ListingsTab } from "@/features/market/listings-tab";
import { MarketFiltersBar } from "@/features/market/market-filters";
import { PurchaseToolDialog, type PurchaseTarget } from "@/features/market/purchase-tool-dialog";
import { ShopTab } from "@/features/market/shop-tab";
import { useMarketFilters } from "@/features/market/use-market-filters";
import { formatCoinAmount } from "@/utils/money/format-coin-amount";

type MarketTab = "shop" | "players";

const marketTabQueryParameter = "tab";

const tabTriggerClassName =
  "rounded-lg px-4 py-2 text-sm font-medium text-muted transition-colors hover:text-foreground data-[state=active]:bg-accent/10 data-[state=active]:text-accent-soft";

function readMarketTab(tabValue: string | null): MarketTab {
  return tabValue === "players" ? "players" : "shop";
}

export function MarketView() {
  const router = useRouter();
  const pathname = usePathname();
  const activeTab = readMarketTab(useSearchParams().get(marketTabQueryParameter));
  const { filters, appliedFilters, changeFilters } = useMarketFilters();
  const [purchaseTarget, setPurchaseTarget] = useState<PurchaseTarget | null>(null);
  const fighterQuery = useFighterProfile();

  function selectTab(tabValue: string) {
    const query = new URLSearchParams({ [marketTabQueryParameter]: readMarketTab(tabValue) });
    router.replace(`${pathname}?${query.toString()}`, { scroll: false });
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="font-display text-4xl tracking-[0.06em] text-foreground">Market</h1>
          <p className="mt-1 text-sm text-muted">Buy new tools from the shop, or grown ones from other fighters.</p>
        </div>
        <span className="flex items-center gap-2 rounded-lg border border-border-strong bg-surface-raised px-3 py-2 text-sm">
          <Coins className="size-4 text-amber-400" aria-hidden="true" />
          <span className="font-mono text-foreground tabular-nums">
            {formatCoinAmount(fighterQuery.data?.coins ?? 0)}
          </span>
          <span className="text-xs text-subtle">coins to spend</span>
        </span>
      </div>

      <Tabs.Root value={activeTab} onValueChange={selectTab} className="space-y-5">
        <Tabs.List aria-label="Market sections" className="flex gap-1 border-b border-border pb-2">
          <Tabs.Trigger value="shop" className={tabTriggerClassName}>
            Shop
          </Tabs.Trigger>
          <Tabs.Trigger value="players" className={tabTriggerClassName}>
            Players
          </Tabs.Trigger>
        </Tabs.List>
        <MarketFiltersBar filters={filters} onChange={changeFilters} showsSort={activeTab === "players"} />
        <Tabs.Content value="shop">
          <ShopTab filters={appliedFilters} onBuy={(shopItem) => setPurchaseTarget({ kind: "SHOP", shopItem })} />
        </Tabs.Content>
        <Tabs.Content value="players">
          <ListingsTab filters={appliedFilters} onBuy={(listing) => setPurchaseTarget({ kind: "LISTING", listing })} />
        </Tabs.Content>
      </Tabs.Root>

      <PurchaseToolDialog purchaseTarget={purchaseTarget} onClose={() => setPurchaseTarget(null)} />
    </div>
  );
}
