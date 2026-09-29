import { http } from "msw";

import type { SignedUpAccount, SignupRequest } from "@/features/auth/api/auth-types";
import { buildApiUrl } from "@/lib/api/api-config";
import { mockDataResponse, mockErrorResponse, simulateLatency } from "@/mocks/utils/mock-responses";

const takenUsernames = new Set(["pilot_nova", "admin", "nebula"]);
const takenEmails = new Set(["taken@nebula.test"]);

export const authHandlers = [
  http.post(buildApiUrl("/auth/signup"), async ({ request }) => {
    await simulateLatency(700);
    const signupRequest = (await request.json()) as SignupRequest;

    const fieldErrors: Record<string, string> = {};
    if (takenEmails.has(signupRequest.email.toLowerCase())) {
      fieldErrors.email = "is already registered";
    }
    if (takenUsernames.has(signupRequest.username.toLowerCase())) {
      fieldErrors.username = "is already taken";
    }
    if (Object.keys(fieldErrors).length > 0) {
      return mockErrorResponse(422, "VALIDATION_FAILED", "Some fields are invalid", fieldErrors);
    }

    const signedUpAccount: SignedUpAccount = {
      id: crypto.randomUUID(),
      email: signupRequest.email,
      username: signupRequest.username,
      status: "UNVERIFIED",
      verification_code_expires_at: new Date(Date.now() + 15 * 60_000).toISOString(),
    };
    return mockDataResponse(signedUpAccount, 201);
  }),
];
