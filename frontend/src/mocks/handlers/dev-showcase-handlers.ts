import { http } from "msw";

import { buildApiUrl } from "@/lib/api/api-config";
import { mockErrorResponse, mockListResponse, simulateLatency } from "@/mocks/utils/mock-responses";

export type SampleTrade = {
  id: string;
  symbol: string;
  side: "buy" | "sell";
  price: string;
  quantity: number;
};

const sampleSymbols = ["IRON/NC", "COPPER/NC", "CRYSTAL/NC", "FUEL/NC"];

const sampleTrades: SampleTrade[] = Array.from({ length: 23 }, (_, tradeIndex) => ({
  id: `trade-${tradeIndex + 1}`,
  symbol: sampleSymbols[tradeIndex % sampleSymbols.length],
  side: tradeIndex % 3 === 0 ? "sell" : "buy",
  price: String(10_000 + tradeIndex * 137),
  quantity: 5 + ((tradeIndex * 7) % 40),
}));

export const devShowcaseHandlers = [
  http.get(buildApiUrl("/dev/sample-trades"), async ({ request }) => {
    await simulateLatency();
    return mockListResponse(sampleTrades, new URL(request.url), 5);
  }),
  http.post(buildApiUrl("/dev/sample-orders"), async () => {
    await simulateLatency();
    return mockErrorResponse(422, "INSUFFICIENT_FUNDS", "You need 0.42 NC more to place this order.", {
      required: "4200000",
      available: "3780000",
    });
  }),
];
