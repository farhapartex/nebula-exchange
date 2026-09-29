import type { UserProfile } from "@/features/auth/api/auth-types";
import { requestData, requestList } from "@/lib/api/api-client";
import type { PaginationParameters } from "@/lib/api/api-types";

export type ActiveSession = {
  id: string;
  user_agent: string;
  ip_address: string;
  started_at: string;
  last_active_at: string;
  is_current: boolean;
};

export type PasswordChangeResult = {
  password_changed: boolean;
  signed_out_other_sessions: number;
};

export function updateUsername(username: string): Promise<UserProfile> {
  return requestData<UserProfile>("/me", { method: "PATCH", body: { username } });
}

export function changePassword(currentPassword: string, newPassword: string): Promise<PasswordChangeResult> {
  return requestData<PasswordChangeResult>("/auth/password-changes", {
    method: "POST",
    body: { current_password: currentPassword, new_password: newPassword },
  });
}

export function listActiveSessions(pagination: PaginationParameters) {
  return requestList<ActiveSession>("/auth/sessions", pagination);
}

export function revokeSession(sessionID: string): Promise<{ revoked: boolean; was_current_session: boolean }> {
  return requestData<{ revoked: boolean; was_current_session: boolean }>(`/auth/sessions/${sessionID}`, {
    method: "DELETE",
  });
}
