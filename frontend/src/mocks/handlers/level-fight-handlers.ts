import { http } from "msw";

import type { FightSetup } from "@/features/fight/api/fight-setup-api";
import { buildApiUrl } from "@/lib/api/api-config";
import levelOneFight from "@/mocks/fixtures/level-fights/1-1.json";
import trainingFight from "@/mocks/fixtures/level-fights/training.json";
import { mockDataResponse, mockErrorResponse, simulateLatency } from "@/mocks/utils/mock-responses";

const fightSetupsByLevel: Record<string, FightSetup> = {
  "1-1": levelOneFight as FightSetup,
  training: trainingFight as FightSetup,
};

export const levelFightHandlers = [
  http.get(buildApiUrl("/levels/:levelID/fight"), async ({ params }) => {
    await simulateLatency();
    const fightSetup = fightSetupsByLevel[String(params.levelID)];
    return fightSetup
      ? mockDataResponse(fightSetup)
      : mockErrorResponse(404, "NOT_FOUND", "This fight is not ready yet");
  }),
];
