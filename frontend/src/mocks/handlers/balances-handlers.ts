import { http } from "msw";

import type { BalanceSummary } from "@/features/balances/api/balances-api";
import { buildApiUrl } from "@/lib/api/api-config";
import { mockDataResponse, simulateLatency } from "@/mocks/utils/mock-responses";

export const mockBalanceSummary: BalanceSummary = {
  total: "31284500",
  available: "28784500",
  held: "2500000",
  withdrawable: "10034500",
  buckets: [
    { bucket: "card", available: "12000000", held: "2500000" },
    { bucket: "earned_pending", available: "6750000", held: "0" },
    { bucket: "earned", available: "4284500", held: "0" },
    { bucket: "crypto", available: "5750000", held: "0" },
  ],
};

export const balancesHandlers = [
  http.get(buildApiUrl("/me/balances"), async () => {
    await simulateLatency();
    return mockDataResponse(mockBalanceSummary);
  }),
];
