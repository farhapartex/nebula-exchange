import { http } from "msw";

import { buildApiUrl } from "@/lib/api/api-config";
import { mockDataResponse, mockErrorResponse, simulateLatency } from "@/mocks/utils/mock-responses";

export const demoRateLimitedEmail = "limited@streetborn.test";

export const activationEmailHandlers = [
  http.post(buildApiUrl("/auth/activation-emails"), async ({ request }) => {
    await simulateLatency(700);
    const { email } = (await request.json()) as { email: string };
    if (email.toLowerCase() === demoRateLimitedEmail) {
      return mockErrorResponse(429, "RATE_LIMITED", "Too many attempts. Please wait and try again.", {
        retry_after_seconds: 60,
      });
    }
    return mockDataResponse({ accepted: true }, 202);
  }),
];
