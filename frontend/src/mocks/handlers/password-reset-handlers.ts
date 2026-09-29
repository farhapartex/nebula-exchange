import { http } from "msw";

import { buildApiUrl } from "@/lib/api/api-config";
import { mockDataResponse, mockErrorResponse, simulateLatency } from "@/mocks/utils/mock-responses";

export const demoPasswordResetToken = "demo-valid-password-reset-token";

const usedPasswordResetTokens = new Set<string>();

function invalidResetLinkResponse() {
  return mockErrorResponse(404, "NOT_FOUND", "This password reset link is invalid or has expired");
}

function isUsableResetToken(resetToken: string): boolean {
  return resetToken === demoPasswordResetToken && !usedPasswordResetTokens.has(resetToken);
}

export const passwordResetHandlers = [
  http.post(buildApiUrl("/auth/password-reset-requests"), async ({ request }) => {
    await simulateLatency(700);
    const { email } = (await request.json()) as { email: string };
    if (email.toLowerCase() === "limited@nebula.test") {
      return mockErrorResponse(429, "RATE_LIMITED", "Too many attempts. Please wait and try again.", {
        retry_after_seconds: 60,
      });
    }
    return mockDataResponse({ accepted: true }, 202);
  }),

  http.get(buildApiUrl("/auth/password-resets/:resetToken"), async ({ params }) => {
    await simulateLatency(500);
    return isUsableResetToken(String(params.resetToken))
      ? mockDataResponse({ email_hint: "p•••t@nebula.test" })
      : invalidResetLinkResponse();
  }),

  http.post(buildApiUrl("/auth/password-resets"), async ({ request }) => {
    await simulateLatency(900);
    const { token, password } = (await request.json()) as { token: string; password: string };
    if (!isUsableResetToken(token)) {
      return invalidResetLinkResponse();
    }
    if (password.toLowerCase().includes("password")) {
      return mockErrorResponse(422, "VALIDATION_FAILED", "Some fields are invalid", {
        password: "is too common",
      });
    }
    usedPasswordResetTokens.add(token);
    return mockDataResponse({ password_reset: true });
  }),
];
