import { http } from "msw";

import type { LevelIntro } from "@/features/level-intro/api/level-intro-api";
import { buildApiUrl } from "@/lib/api/api-config";
import levelOneIntro from "@/mocks/fixtures/level-intros/1-1.json";
import { mockDataResponse, mockErrorResponse, simulateLatency } from "@/mocks/utils/mock-responses";

const levelIntrosByID: Record<string, LevelIntro> = {
  "1-1": levelOneIntro as LevelIntro,
};

export const levelIntroHandlers = [
  http.get(buildApiUrl("/levels/:levelID/intro"), async ({ params }) => {
    await simulateLatency();
    const levelIntro = levelIntrosByID[String(params.levelID)];
    return levelIntro
      ? mockDataResponse(levelIntro)
      : mockErrorResponse(404, "NOT_FOUND", "This level has no story yet");
  }),
];
