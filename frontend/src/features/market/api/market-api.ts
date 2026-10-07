import { requestData, requestList } from "@/lib/api/api-client";
import type { MarketFilters, MarketListing, ShopItem } from "@/features/market/api/market-types";

const listingsPageSize = 12;
const shopPageSize = 20;

export const marketQueryKeys = {
  shopItems: (filters: MarketFilters) => ["shop-items", filters.category, filters.rarity, filters.search] as const,
  listings: (filters: MarketFilters) =>
    ["market-listings", filters.category, filters.rarity, filters.sort, filters.search] as const,
  allShopItems: ["shop-items"] as const,
  allListings: ["market-listings"] as const,
};

function filterQuery(filters: MarketFilters) {
  return {
    category: filters.category === "ALL" ? undefined : filters.category,
    rarity: filters.rarity === "ALL" ? undefined : filters.rarity,
    q: filters.search.trim() || undefined,
  };
}

export function fetchShopItemsPage(filters: MarketFilters, cursor: string | null) {
  return requestList<ShopItem>("/shop-items", { cursor, limit: shopPageSize }, { query: filterQuery(filters) });
}

export function fetchMarketListingsPage(filters: MarketFilters, cursor: string | null) {
  return requestList<MarketListing>(
    "/market-listings",
    { cursor, limit: listingsPageSize },
    { query: { ...filterQuery(filters), sort: filters.sort } },
  );
}

export function buyShopItem(toolTypeID: string) {
  return requestData<{ tool_id: string }>(`/shop-items/${encodeURIComponent(toolTypeID)}/purchases`, {
    method: "POST",
    body: {},
  });
}

export function buyMarketListing(listingID: string) {
  return requestData<{ tool_id: string }>(`/market-listings/${encodeURIComponent(listingID)}/purchases`, {
    method: "POST",
    body: {},
  });
}
