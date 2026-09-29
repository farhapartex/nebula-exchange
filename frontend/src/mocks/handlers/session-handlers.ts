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

const mockLockoutFailureLimit = 5;
const mockLockoutDurationInSeconds = 15 * 60;
const failedLoginAttemptsByEmail = new Map<string, number>();
const lockedUntilByEmail = new Map<string, number>();

function lockedResponseFor(email: string) {
  const lockedUntil = lockedUntilByEmail.get(email);
  if (!lockedUntil || lockedUntil <= Date.now()) {
    return null;
  }
  return mockErrorResponse(429, "LOGIN_LOCKED", "Too many failed login attempts. Try again later.", {
    retry_after_seconds: Math.ceil((lockedUntil - Date.now()) / 1000),
  });
}

function recordFailedLogin(email: string) {
  const failedAttempts = (failedLoginAttemptsByEmail.get(email) ?? 0) + 1;
  failedLoginAttemptsByEmail.set(email, failedAttempts);
  if (failedAttempts >= mockLockoutFailureLimit) {
    lockedUntilByEmail.set(email, Date.now() + mockLockoutDurationInSeconds * 1000);
    failedLoginAttemptsByEmail.delete(email);
  }
}

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
    const normalizedEmail = loginRequest.email.toLowerCase();
    if (normalizedEmail === "busy@nebula.test") {
      return mockErrorResponse(429, "RATE_LIMITED", "Too many attempts. Please wait and try again.", {
        retry_after_seconds: 45,
      });
    }
    const lockedResponse = lockedResponseFor(normalizedEmail);
    if (lockedResponse) {
      return lockedResponse;
    }
    const mockAccount = mockAccounts.find((account) => account.email === normalizedEmail);

    if (!mockAccount || loginRequest.password !== mockPassword) {
      recordFailedLogin(normalizedEmail);
      return lockedResponseFor(normalizedEmail) ?? mockErrorResponse(401, "UNAUTHORIZED", "Invalid email or password");
    }
    failedLoginAttemptsByEmail.delete(normalizedEmail);
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

  http.post(buildApiUrl("/auth/logout"), async () => {
    await simulateLatency(300);
    signedInMockAccount = null;
    return mockDataResponse({ logged_out: true });
  }),

  http.post(buildApiUrl("/dev/mock-session/expire"), async () => {
    signedInMockAccount = null;
    return mockDataResponse({ expired: true });
  }),

  http.get(buildApiUrl("/me"), async () => {
    await simulateLatency(200);
    if (!signedInMockAccount) {
      return mockErrorResponse(401, "UNAUTHORIZED", "Log in to continue");
    }
    return mockDataResponse(toProfile(signedInMockAccount));
  }),
];
