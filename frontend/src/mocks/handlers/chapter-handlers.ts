import { http } from "msw";

import type { PurchasableChapter } from "@/features/chapter-purchase/api/chapter-catalog-api";
import { buildApiUrl } from "@/lib/api/api-config";
import purchasableChapters from "@/mocks/fixtures/chapters/purchasable-chapters.json";
import { mockListResponse, simulateLatency } from "@/mocks/utils/mock-responses";

export const chapterHandlers = [
  http.get(buildApiUrl("/chapters"), async ({ request }) => {
    await simulateLatency();
    return mockListResponse(purchasableChapters as PurchasableChapter[], new URL(request.url));
  }),
];
