import { http } from "msw";

import type { ActiveSession } from "@/features/settings/api/settings-api";
import { buildApiUrl } from "@/lib/api/api-config";
import { mockDataResponse, mockErrorResponse, mockListResponse, simulateLatency } from "@/mocks/utils/mock-responses";

function hoursAgo(hours: number) {
  return new Date(Date.now() - hours * 3_600_000).toISOString();
}

let mockSessions: ActiveSession[] = [
  {
    id: "session-current",
    user_agent: navigatorUserAgent(),
    ip_address: "203.0.113.5",
    started_at: hoursAgo(26),
    last_active_at: hoursAgo(0.01),
    is_current: true,
  },
  {
    id: "session-phone",
    user_agent:
      "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 Version/18.0 Mobile Safari/604.1",
    ip_address: "198.51.100.23",
    started_at: hoursAgo(120),
    last_active_at: hoursAgo(3),
    is_current: false,
  },
  {
    id: "session-laptop",
    user_agent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0",
    ip_address: "192.0.2.77",
    started_at: hoursAgo(300),
    last_active_at: hoursAgo(48),
    is_current: false,
  },
];

function navigatorUserAgent(): string {
  return typeof navigator === "undefined" ? "" : navigator.userAgent;
}

export const settingsHandlers = [
  http.patch(buildApiUrl("/me"), async ({ request }) => {
    await simulateLatency(500);
    const { username } = (await request.json()) as { username: string };
    if (["pilot_nova", "admin"].includes(username.toLowerCase())) {
      return mockErrorResponse(422, "VALIDATION_FAILED", "Some fields are invalid", { username: "is already taken" });
    }
    return mockDataResponse({
      id: "01a0edce-0000-7000-8000-000000000001",
      email: "pilot@nebula.test",
      username,
      status: "ACTIVE",
      is_active: true,
      is_admin: false,
      two_factor_enabled: false,
      created_at: "2026-09-29T10:00:00Z",
      last_login_at: new Date().toISOString(),
    });
  }),

  http.post(buildApiUrl("/auth/password-changes"), async ({ request }) => {
    await simulateLatency(700);
    const { current_password: currentPassword } = (await request.json()) as { current_password: string };
    if (currentPassword !== "Mining4Crystal!Moon") {
      return mockErrorResponse(422, "VALIDATION_FAILED", "Some fields are invalid", {
        current_password: "is incorrect",
      });
    }
    const signedOutCount = mockSessions.filter((mockSession) => !mockSession.is_current).length;
    mockSessions = mockSessions.filter((mockSession) => mockSession.is_current);
    return mockDataResponse({ password_changed: true, signed_out_other_sessions: signedOutCount });
  }),

  http.get(buildApiUrl("/auth/sessions"), async ({ request }) => {
    await simulateLatency(400);
    return mockListResponse(mockSessions, new URL(request.url), 10);
  }),

  http.delete(buildApiUrl("/auth/sessions/:sessionID"), async ({ params }) => {
    await simulateLatency(400);
    const revokedSession = mockSessions.find((mockSession) => mockSession.id === params.sessionID);
    if (!revokedSession) {
      return mockErrorResponse(404, "NOT_FOUND", "This session no longer exists");
    }
    mockSessions = mockSessions.filter((mockSession) => mockSession.id !== params.sessionID);
    return mockDataResponse({ revoked: true, was_current_session: revokedSession.is_current });
  }),
];
