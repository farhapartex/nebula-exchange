import { http } from "msw";

import { buildApiUrl } from "@/lib/api/api-config";
import {
  mockCatalogItems,
  mockRecipes,
  mockShopItems,
  mockUpgrades,
  mockZones,
} from "@/mocks/fixtures/catalog-fixtures";
import { buildMockItemUsage } from "@/mocks/fixtures/item-usage";
import { mockDataResponse, mockErrorResponse, mockListResponse, simulateLatency } from "@/mocks/utils/mock-responses";

function catalogListHandler<Entry>(path: string, entries: Entry[]) {
  return http.get(buildApiUrl(path), async ({ request }) => {
    await simulateLatency();
    return mockListResponse(entries, new URL(request.url));
  });
}

export const catalogHandlers = [
  catalogListHandler("/items", mockCatalogItems),
  http.get(buildApiUrl("/items/:itemID"), async ({ params }) => {
    await simulateLatency();
    const catalogItem = mockCatalogItems.find((candidate) => String(candidate.id) === params.itemID);
    return catalogItem
      ? mockDataResponse({ ...catalogItem, usage: buildMockItemUsage(catalogItem.id) })
      : mockErrorResponse(404, "NOT_FOUND", "This item does not exist");
  }),
  catalogListHandler("/recipes", mockRecipes),
  catalogListHandler("/upgrades", mockUpgrades),
  catalogListHandler("/zones", mockZones),
  catalogListHandler("/shop-items", mockShopItems),
];
