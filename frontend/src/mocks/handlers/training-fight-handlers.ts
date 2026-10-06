import { http, passthrough } from "msw";

import type { FightSetup } from "@/features/fight/api/fight-setup-api";
import { buildApiUrl } from "@/lib/api/api-config";
import trainingFight from "@/mocks/fixtures/level-fights/training.json";
import { mockDataResponse, simulateLatency } from "@/mocks/utils/mock-responses";

const trainingLevelID = "training";

export const trainingFightHandlers = [
  http.get(buildApiUrl("/levels/:levelID/fight-setup"), async ({ params }) => {
    if (String(params.levelID) !== trainingLevelID) {
      return passthrough();
    }
    await simulateLatency();
    return mockDataResponse(trainingFight as FightSetup);
  }),
];
