import { http } from "msw";

import type { InventoryHolding } from "@/features/inventory/api/inventory-api";
import { buildApiUrl } from "@/lib/api/api-config";
import { mockListResponse, simulateLatency } from "@/mocks/utils/mock-responses";

export const mockInventory: InventoryHolding[] = [
  { item_id: 1, available: 84, held: 20 },
  { item_id: 2, available: 31, held: 0 },
  { item_id: 3, available: 12, held: 0 },
  { item_id: 101, available: 6, held: 0 },
  { item_id: 102, available: 2, held: 0 },
  { item_id: 201, available: 0, held: 1 },
  { item_id: 202, available: 1, held: 0 },
  { item_id: 301, available: 1, held: 0 },
  { item_id: 302, available: 1, held: 0 },
  { item_id: 401, available: 14, held: 0 },
];

export const inventoryHandlers = [
  http.get(buildApiUrl("/me/inventory"), async ({ request }) => {
    await simulateLatency();
    return mockListResponse(mockInventory, new URL(request.url));
  }),
];
