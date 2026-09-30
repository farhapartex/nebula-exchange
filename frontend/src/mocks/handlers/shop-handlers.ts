import { http } from "msw";

import type { CompletedPurchase } from "@/features/shop/api/shop-api";
import { buildApiUrl } from "@/lib/api/api-config";
import { mockShopItems } from "@/mocks/fixtures/catalog-fixtures";
import { mockBalanceSummary } from "@/mocks/handlers/balances-handlers";
import { mockDataResponse, mockErrorResponse, simulateLatency } from "@/mocks/utils/mock-responses";

export const shopHandlers = [
  http.post(buildApiUrl("/shop-items/:sku/purchases"), async ({ params, request }) => {
    await simulateLatency();
    const shopItem = mockShopItems.find((candidate) => candidate.sku === params.sku);
    if (!shopItem) {
      return mockErrorResponse(404, "NOT_FOUND", "This item is not sold in the shop");
    }
    const { quantity } = (await request.json()) as { quantity: number };
    const total = BigInt(shopItem.price) * BigInt(quantity);
    if (total > BigInt(mockBalanceSummary.available)) {
      return mockErrorResponse(409, "INSUFFICIENT_FUNDS", "You don't have enough NC for this");
    }
    const completedPurchase: CompletedPurchase = {
      id: crypto.randomUUID(),
      sku: shopItem.sku,
      quantity,
      total: total.toString(),
      journal_id: crypto.randomUUID(),
    };
    return mockDataResponse(completedPurchase, 201);
  }),
];
