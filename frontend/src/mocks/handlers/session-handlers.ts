import { http } from "msw";

import type { EstablishedSession, LoginRequest, UserProfile } from "@/features/auth/api/auth-types";
import { buildApiUrl } from "@/lib/api/api-config";
import { mockDataResponse, mockErrorResponse, simulateLatency } from "@/mocks/utils/mock-responses";

const mockPassword = "Mining4Crystal!Moon";

type MockAccount = UserProfile & { outcome: "success" | "not_activated" | "banned" };

const mockAccounts: MockAccount[] = [
  {
    id: "01a0edce-0000-7000-8000-000000000001",
    email: "pilot@nebula.test",
    username: "pilot_nova",
    status: "PENDING_PAYMENT",
    is_active: true,
    is_admin: false,
    created_at: "2026-09-29T10:00:00Z",
    last_login_at: null,
    outcome: "success",
  },
  {
    id: "01a0edce-0000-7000-8000-000000000002",
    email: "inactive@nebula.test",
    username: "sleepy_miner",
    status: "UNVERIFIED",
    is_active: false,
    is_admin: false,
    created_at: "2026-09-29T10:00:00Z",
    last_login_at: null,
    outcome: "not_activated",
  },
  {
    id: "01a0edce-0000-7000-8000-000000000003",
    email: "banned@nebula.test",
    username: "space_pirate",
    status: "BANNED",
    is_active: true,
    is_admin: false,
    created_at: "2026-09-29T10:00:00Z",
    last_login_at: null,
    outcome: "banned",
  },
];

let signedInMockAccount: MockAccount | null = null;

function toProfile(mockAccount: MockAccount): UserProfile {
  const profile: Partial<MockAccount> = { ...mockAccount };
  delete profile.outcome;
  return profile as UserProfile;
}

function buildMockSession(mockAccount: MockAccount): EstablishedSession {
  return {
    access_token: `mock-access-token-${mockAccount.id}-${Date.now()}`,
    access_token_expires_at: new Date(Date.now() + 15 * 60_000).toISOString(),
    user: toProfile(mockAccount),
  };
}

export const sessionHandlers = [
  http.post(buildApiUrl("/auth/login"), async ({ request }) => {
    await simulateLatency(600);
    const loginRequest = (await request.json()) as LoginRequest;
    const mockAccount = mockAccounts.find((account) => account.email === loginRequest.email.toLowerCase());

    if (!mockAccount || loginRequest.password !== mockPassword) {
      return mockErrorResponse(401, "UNAUTHORIZED", "Invalid email or password");
    }
    if (mockAccount.outcome === "not_activated") {
      return mockErrorResponse(
        403,
        "ACCOUNT_NOT_ACTIVATED",
        "Your account is not activated yet. Check your email for the activation link.",
      );
    }
    if (mockAccount.outcome === "banned") {
      return mockErrorResponse(403, "FORBIDDEN", "This account is banned");
    }

    mockAccount.last_login_at = new Date().toISOString();
    signedInMockAccount = mockAccount;
    return mockDataResponse(buildMockSession(mockAccount));
  }),

  http.post(buildApiUrl("/auth/refresh"), async () => {
    await simulateLatency(200);
    if (!signedInMockAccount) {
      return mockErrorResponse(401, "UNAUTHORIZED", "Your session has expired. Log in again.");
    }
    return mockDataResponse(buildMockSession(signedInMockAccount));
  }),

  http.get(buildApiUrl("/me"), async () => {
    await simulateLatency(200);
    if (!signedInMockAccount) {
      return mockErrorResponse(401, "UNAUTHORIZED", "Log in to continue");
    }
    return mockDataResponse(toProfile(signedInMockAccount));
  }),
];
