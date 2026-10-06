"use client";

import { useQuery } from "@tanstack/react-query";

import { ErrorState } from "@/components/ui/error-state";
import { fetchShopItems, marketQueryKeys } from "@/features/market/api/market-api";
import type { MarketFilters, ShopItem } from "@/features/market/api/market-types";
import { MarketEmptyState, MarketGridSkeleton } from "@/features/market/market-grid-states";
import { ShopItemCard } from "@/features/market/shop-item-card";

type ShopTabProps = {
  filters: MarketFilters;
  onBuy: (shopItem: ShopItem) => void;
};

export function ShopTab({ filters, onBuy }: ShopTabProps) {
  const shopQuery = useQuery({ queryKey: marketQueryKeys.shopItems(filters), queryFn: () => fetchShopItems(filters) });

  if (shopQuery.isError) {
    return (
      <ErrorState
        message="The shop could not be loaded. Try again."
        onRetry={() => void shopQuery.refetch()}
        isRetrying={shopQuery.isFetching}
      />
    );
  }
  if (!shopQuery.data) {
    return <MarketGridSkeleton />;
  }
  if (shopQuery.data.length === 0) {
    return <MarketEmptyState message="No tools match these filters." />;
  }
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {shopQuery.data.map((shopItem) => (
        <ShopItemCard key={shopItem.tool_type.id} shopItem={shopItem} onBuy={onBuy} />
      ))}
    </div>
  );
}
