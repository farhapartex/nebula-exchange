import { http } from "msw";

import type { ActivatedAccount, ActivationPreview } from "@/features/auth/api/auth-types";
import { buildApiUrl } from "@/lib/api/api-config";
import { mockDataResponse, mockErrorResponse, simulateLatency } from "@/mocks/utils/mock-responses";

export const demoRateLimitedEmail = "limited@nebula.test";

export const demoActivationTokens = {
  pending: "demo-pending-activation-token",
  alreadyActivated: "demo-already-activated-token",
  expired: "demo-expired-activation-token",
} as const;

const mockActivationPreviews: Record<string, ActivationPreview> = {
  [demoActivationTokens.pending]: { email: "new.pilot@nebula.test", username: "new_pilot", is_activated: false },
  [demoActivationTokens.alreadyActivated]: { email: "pilot@nebula.test", username: "pilot_nova", is_activated: true },
};

function invalidActivationLinkResponse() {
  return mockErrorResponse(404, "NOT_FOUND", "This activation link is invalid or has expired");
}

export const activationHandlers = [
  http.get(buildApiUrl("/auth/activations/:activationToken"), async ({ params }) => {
    await simulateLatency(500);
    const activationPreview = mockActivationPreviews[String(params.activationToken)];
    return activationPreview ? mockDataResponse(activationPreview) : invalidActivationLinkResponse();
  }),

  http.post(buildApiUrl("/auth/activations"), async ({ request }) => {
    await simulateLatency(1200);
    const { token } = (await request.json()) as { token: string };
    const activationPreview = mockActivationPreviews[token];
    if (!activationPreview) {
      return invalidActivationLinkResponse();
    }

    activationPreview.is_activated = true;
    const activatedAccount: ActivatedAccount = {
      email: activationPreview.email,
      username: activationPreview.username,
      status: "PENDING_PAYMENT",
      activated_at: new Date().toISOString(),
    };
    return mockDataResponse(activatedAccount);
  }),

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
