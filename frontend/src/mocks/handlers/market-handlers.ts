import { http } from "msw";

import type { ListingSort, MarketListing, ShopItem } from "@/features/market/api/market-types";
import { buildApiUrl } from "@/lib/api/api-config";
import marketListingsFixture from "@/mocks/fixtures/market/market-listings.json";
import shopItemsFixture from "@/mocks/fixtures/market/shop-items.json";
import { spendMockCoins } from "@/mocks/utils/mock-coin-balance";
import { mockDataResponse, mockErrorResponse, mockListResponse, simulateLatency } from "@/mocks/utils/mock-responses";
import { readSessionMockState, writeSessionMockState } from "@/mocks/utils/session-mock-store";

const mockShopItems = shopItemsFixture as unknown as ShopItem[];
const mockMarketListings = marketListingsFixture as unknown as MarketListing[];

const boughtShopItemsStorageKey = "street-born-mock-bought-shop-items";
const boughtListingsStorageKey = "street-born-mock-bought-listings";

const listingComparators: Record<ListingSort, (first: MarketListing, second: MarketListing) => number> = {
  PRICE_LOW: (first, second) => Number(first.price_coins) - Number(second.price_coins),
  PRICE_HIGH: (first, second) => Number(second.price_coins) - Number(first.price_coins),
  MASTERY_HIGH: (first, second) => second.tool.mastery_level - first.tool.mastery_level,
  NEWEST: (first, second) => second.listed_at.localeCompare(first.listed_at),
};

type ToolTypeFilterSubject = { category: string; rarity: string; name: string };

function matchesFilters(toolType: ToolTypeFilterSubject, searchParameters: URLSearchParams): boolean {
  const category = searchParameters.get("category");
  const rarity = searchParameters.get("rarity");
  const search = searchParameters.get("search")?.toLowerCase();
  return (
    (!category || toolType.category === category) &&
    (!rarity || toolType.rarity === rarity) &&
    (!search || toolType.name.toLowerCase().includes(search))
  );
}

function readBoughtShopItems(): Record<string, number> {
  return readSessionMockState<Record<string, number>>(boughtShopItemsStorageKey, {});
}

function readBoughtListings(): string[] {
  return readSessionMockState<string[]>(boughtListingsStorageKey, []);
}

export const marketHandlers = [
  http.get(buildApiUrl("/shop-items"), async ({ request }) => {
    await simulateLatency();
    const requestUrl = new URL(request.url);
    const boughtShopItems = readBoughtShopItems();
    const shopItems = mockShopItems
      .filter((shopItem) => matchesFilters(shopItem.tool_type, requestUrl.searchParams))
      .map((shopItem) => ({
        ...shopItem,
        owned_count: shopItem.owned_count + (boughtShopItems[shopItem.tool_type.id] ?? 0),
      }));
    return mockListResponse(shopItems, requestUrl, 100);
  }),

  http.get(buildApiUrl("/market-listings"), async ({ request }) => {
    await simulateLatency();
    const requestUrl = new URL(request.url);
    const boughtListings = new Set(readBoughtListings());
    const sort = (requestUrl.searchParams.get("sort") as ListingSort | null) ?? "PRICE_LOW";
    const listings = mockMarketListings
      .filter((listing) => !boughtListings.has(listing.id))
      .filter((listing) => matchesFilters(listing.tool.tool_type, requestUrl.searchParams))
      .sort(listingComparators[sort] ?? listingComparators.PRICE_LOW);
    return mockListResponse(listings, requestUrl, 12);
  }),

  http.post(buildApiUrl("/shop-items/:toolTypeID/purchases"), async ({ params }) => {
    await simulateLatency();
    const shopItem = mockShopItems.find((candidate) => candidate.tool_type.id === String(params.toolTypeID));
    if (!shopItem) {
      return mockErrorResponse(404, "NOT_FOUND", "This tool is not in the shop");
    }
    if (!shopItem.is_unlocked) {
      return mockErrorResponse(403, "FORBIDDEN", "Win the level that unlocks this tool first");
    }
    if (!spendMockCoins(Number(shopItem.price_coins))) {
      return mockErrorResponse(422, "INSUFFICIENT_FUNDS", "You do not have enough coins");
    }
    const boughtShopItems = readBoughtShopItems();
    boughtShopItems[shopItem.tool_type.id] = (boughtShopItems[shopItem.tool_type.id] ?? 0) + 1;
    writeSessionMockState(boughtShopItemsStorageKey, boughtShopItems);
    return mockDataResponse({ tool_id: crypto.randomUUID() }, 201);
  }),

  http.post(buildApiUrl("/market-listings/:listingID/purchases"), async ({ params }) => {
    await simulateLatency();
    const listingID = String(params.listingID);
    const listing = mockMarketListings.find((candidate) => candidate.id === listingID);
    if (!listing || readBoughtListings().includes(listingID)) {
      return mockErrorResponse(404, "NOT_FOUND", "This listing is no longer for sale");
    }
    if (listing.is_own) {
      return mockErrorResponse(409, "CONFLICT", "You cannot buy your own listing");
    }
    if (!spendMockCoins(Number(listing.price_coins))) {
      return mockErrorResponse(422, "INSUFFICIENT_FUNDS", "You do not have enough coins");
    }
    writeSessionMockState(boughtListingsStorageKey, [...readBoughtListings(), listingID]);
    return mockDataResponse({ tool_id: listing.tool.id }, 201);
  }),
];
