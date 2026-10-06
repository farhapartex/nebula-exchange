import { http } from "msw";

import type { FighterProfile, LoadoutSlot, StoryChapter } from "@/features/fight-hub/api/fight-hub-types";
import { buildApiUrl } from "@/lib/api/api-config";
import { readMockCoinBalance } from "@/mocks/utils/mock-coin-balance";
import { mockDataResponse, simulateLatency } from "@/mocks/utils/mock-responses";

const mockFighter: FighterProfile = {
  fighter_level: 1,
  experience: 0,
  experience_to_next_level: 100,
  wins: 0,
  losses: 0,
  coins: 0,
};

const mockStory: StoryChapter[] = [
  {
    id: "chapter-1",
    number: 1,
    title: "The night they came",
    is_free: true,
    is_unlocked: true,
    levels: [
      {
        id: "1-1",
        number: 1,
        title: "The burning house",
        teaser: "Get out alive.",
        status: "current",
        best_stars: null,
      },
      {
        id: "1-2",
        number: 2,
        title: "The hammer",
        teaser: "One hit could end it.",
        status: "locked",
        best_stars: null,
      },
      { id: "1-3", number: 3, title: "The knife", teaser: "He's fast. Be faster.", status: "locked", best_stars: null },
      {
        id: "1-4",
        number: 4,
        title: "The gun",
        teaser: "Close the distance or fall.",
        status: "locked",
        best_stars: null,
      },
      {
        id: "1-5",
        number: 5,
        title: "Last stand",
        teaser: "All of them. At once.",
        status: "locked",
        best_stars: null,
      },
      {
        id: "1-finale",
        number: null,
        title: "The offer",
        teaser: "Someone was watching.",
        status: "locked",
        best_stars: null,
      },
    ],
  },
  { id: "chapter-2", number: 2, title: "The club", is_free: false, is_unlocked: false, levels: [] },
];

const mockLoadout: LoadoutSlot[] = [
  { slot_number: 1, tool: { id: "tool-fists", key: "bare_fists", name: "Bare fists", mastery_percent: 12 } },
  { slot_number: 2, tool: null },
  { slot_number: 3, tool: null },
];

export const fightHubHandlers = [
  http.get(buildApiUrl("/me/fighter"), async () => {
    await simulateLatency();
    return mockDataResponse({ ...mockFighter, coins: readMockCoinBalance() });
  }),
  http.get(buildApiUrl("/me/story"), async () => {
    await simulateLatency();
    return mockDataResponse(mockStory);
  }),
  http.get(buildApiUrl("/me/loadout"), async () => {
    await simulateLatency();
    return mockDataResponse(mockLoadout);
  }),
];
