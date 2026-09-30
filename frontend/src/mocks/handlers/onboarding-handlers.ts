import { http } from "msw";

import type { OnboardingProgress } from "@/features/hangar/api/onboarding-api";
import { buildApiUrl } from "@/lib/api/api-config";
import { mockDataResponse, simulateLatency } from "@/mocks/utils/mock-responses";

const mockOnboardingProgress: OnboardingProgress = {
  steps: [
    { key: "paid_entry_fee", is_completed: true },
    { key: "first_mission", is_completed: true },
    { key: "first_craft", is_completed: false },
    { key: "first_trade", is_completed: false },
  ],
  completed_count: 2,
};

export const onboardingHandlers = [
  http.get(buildApiUrl("/me/onboarding"), async () => {
    await simulateLatency();
    return mockDataResponse(mockOnboardingProgress);
  }),
];
