import { http } from "msw";

import { buildApiUrl } from "@/lib/api/api-config";
import { mockDataResponse, simulateLatency } from "@/mocks/utils/mock-responses";

export const healthHandlers = [
  http.get(buildApiUrl("/health"), async () => {
    await simulateLatency();
    return mockDataResponse({ status: "ok", checked_at: new Date().toISOString() });
  }),
];
