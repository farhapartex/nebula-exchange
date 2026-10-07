"use client";

import { useInfiniteQuery } from "@tanstack/react-query";

import { Button } from "@/components/ui/button";
import { ErrorState } from "@/components/ui/error-state";
import { useCurrentPlayer } from "@/features/auth/session/use-current-player";
import { fetchMarketListingsPage, marketQueryKeys } from "@/features/market/api/market-api";
import type { MarketFilters, MarketListing } from "@/features/market/api/market-types";
import { levelsStillNeeded } from "@/features/market/fighter-level-requirement";
import { ListingCard } from "@/features/market/listing-card";
import { MarketEmptyState, MarketGridSkeleton } from "@/features/market/market-grid-states";

type ListingsTabProps = {
  filters: MarketFilters;
  onBuy: (listing: MarketListing) => void;
};

export function ListingsTab({ filters, onBuy }: ListingsTabProps) {
  const currentFighterLevel = useCurrentPlayer().data?.current_level;
  const listingsQuery = useInfiniteQuery({
    queryKey: marketQueryKeys.listings(filters),
    queryFn: ({ pageParam }) => fetchMarketListingsPage(filters, pageParam),
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => lastPage.pagination.next_cursor,
  });

  if (listingsQuery.isError) {
    return (
      <ErrorState
        message="The player market could not be loaded. Try again."
        onRetry={() => void listingsQuery.refetch()}
        isRetrying={listingsQuery.isFetching}
      />
    );
  }
  if (!listingsQuery.data) {
    return <MarketGridSkeleton />;
  }
  const listings = listingsQuery.data.pages.flatMap((listingsPage) => listingsPage.data);
  if (listings.length === 0) {
    return <MarketEmptyState message="Nobody is selling a tool like this right now." />;
  }
  return (
    <div className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {listings.map((listing) => (
          <ListingCard
            key={listing.id}
            listing={listing}
            levelsNeeded={levelsStillNeeded(listing.tool.tool_type.minimum_fighter_level, currentFighterLevel)}
            onBuy={onBuy}
          />
        ))}
      </div>
      {listingsQuery.hasNextPage && (
        <div className="flex justify-center">
          <Button
            variant="secondary"
            onClick={() => void listingsQuery.fetchNextPage()}
            isLoading={listingsQuery.isFetchingNextPage}
          >
            Show more
          </Button>
        </div>
      )}
    </div>
  );
}
