"use client";

import { useInfiniteQuery } from "@tanstack/react-query";

import { Button } from "@/components/ui/button";
import { ErrorState } from "@/components/ui/error-state";
import { useCurrentPlayer } from "@/features/auth/session/use-current-player";
import { fetchShopItemsPage, marketQueryKeys } from "@/features/market/api/market-api";
import type { MarketFilters, ShopItem } from "@/features/market/api/market-types";
import { MarketEmptyState, MarketGridSkeleton } from "@/features/market/market-grid-states";
import { levelsStillNeeded } from "@/features/market/fighter-level-requirement";
import { ShopItemCard } from "@/features/market/shop-item-card";

type ShopTabProps = {
  filters: MarketFilters;
  onBuy: (shopItem: ShopItem) => void;
};

export function ShopTab({ filters, onBuy }: ShopTabProps) {
  const currentFighterLevel = useCurrentPlayer().data?.current_level;
  const shopQuery = useInfiniteQuery({
    queryKey: marketQueryKeys.shopItems(filters),
    queryFn: ({ pageParam }) => fetchShopItemsPage(filters, pageParam),
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => lastPage.pagination.next_cursor,
  });

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
  const shopItems = shopQuery.data.pages.flatMap((shopPage) => shopPage.data);
  if (shopItems.length === 0) {
    return <MarketEmptyState message="No tools match these filters." />;
  }
  return (
    <div className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {shopItems.map((shopItem) => (
          <ShopItemCard
            key={shopItem.tool_type.id}
            shopItem={shopItem}
            levelsNeeded={
              shopItem.is_usable
                ? 0
                : Math.max(levelsStillNeeded(shopItem.tool_type.minimum_fighter_level, currentFighterLevel), 1)
            }
            onBuy={onBuy}
          />
        ))}
      </div>
      {shopQuery.hasNextPage && (
        <div className="flex justify-center">
          <Button
            variant="secondary"
            onClick={() => void shopQuery.fetchNextPage()}
            isLoading={shopQuery.isFetchingNextPage}
          >
            Show more
          </Button>
        </div>
      )}
    </div>
  );
}
