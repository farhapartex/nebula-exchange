import { http } from "msw";

import type { LedgerJournal } from "@/features/history/api/ledger-history-api";
import { buildApiUrl } from "@/lib/api/api-config";
import { mockListResponse, simulateLatency } from "@/mocks/utils/mock-responses";

function minutesAgo(minutes: number) {
  return new Date(Date.now() - minutes * 60_000).toISOString();
}

const mockLedgerJournals: LedgerJournal[] = [
  {
    id: "01a0f000-0000-7000-8000-000000000009",
    type: "TRADE_FILL",
    reference: { type: "trade", id: "7f1c2d9e-trade" },
    created_at: minutesAgo(4),
    entries: [
      { item_id: null, bucket: "earned_pending", amount: "3900000" },
      { item_id: 1, bucket: null, amount: "-20" },
    ],
  },
  {
    id: "01a0f000-0000-7000-8000-000000000008",
    type: "MISSION_LOOT",
    reference: { type: "mission", id: "a81b3c44-mission" },
    created_at: minutesAgo(38),
    entries: [
      { item_id: 1, bucket: null, amount: "12" },
      { item_id: 2, bucket: null, amount: "5" },
    ],
  },
  {
    id: "01a0f000-0000-7000-8000-000000000007",
    type: "MISSION_FUEL",
    reference: { type: "mission", id: "a81b3c44-mission" },
    created_at: minutesAgo(53),
    entries: [{ item_id: 401, bucket: null, amount: "-2" }],
  },
  {
    id: "01a0f000-0000-7000-8000-000000000006",
    type: "CRAFT_START",
    reference: { type: "craft", id: "c93d1e55-craft" },
    created_at: minutesAgo(90),
    entries: [
      { item_id: null, bucket: "card", amount: "-20000" },
      { item_id: 1, bucket: null, amount: "-5" },
      { item_id: 2, bucket: null, amount: "-2" },
    ],
  },
  {
    id: "01a0f000-0000-7000-8000-000000000005",
    type: "STARTER_PACK",
    reference: { type: "user", id: "starter-pack" },
    created_at: minutesAgo(60 * 26),
    entries: [
      { item_id: null, bucket: "card", amount: "2000000" },
      { item_id: 201, bucket: null, amount: "1" },
      { item_id: 301, bucket: null, amount: "1" },
      { item_id: 401, bucket: null, amount: "10" },
    ],
  },
  {
    id: "01a0f000-0000-7000-8000-000000000004",
    type: "ENTRY_FEE",
    reference: { type: "payment", id: "5be2f8a1-payment" },
    created_at: minutesAgo(60 * 26 + 1),
    entries: [{ item_id: null, bucket: "card", amount: "-5000000" }],
  },
  {
    id: "01a0f000-0000-7000-8000-000000000003",
    type: "TOPUP_CARD",
    reference: { type: "payment", id: "5be2f8a1-payment" },
    created_at: minutesAgo(60 * 26 + 2),
    entries: [{ item_id: null, bucket: "card", amount: "5000000" }],
  },
];

export const ledgerHistoryHandlers = [
  http.get(buildApiUrl("/me/ledger"), async ({ request }) => {
    await simulateLatency();
    const requestUrl = new URL(request.url);
    const journalType = requestUrl.searchParams.get("type");
    const matchingJournals = journalType
      ? mockLedgerJournals.filter((journal) => journal.type === journalType)
      : mockLedgerJournals;
    return mockListResponse(matchingJournals, requestUrl);
  }),
];
