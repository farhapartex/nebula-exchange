import { requestData } from "@/lib/api/api-client";

export type PasswordResetPreview = {
  email_hint: string;
};

export type PasswordResetResult = {
  password_reset: boolean;
};

export function requestPasswordReset(emailAddress: string): Promise<{ accepted: boolean }> {
  return requestData<{ accepted: boolean }>("/auth/password-reset-requests", {
    method: "POST",
    body: { email: emailAddress },
    skipSessionRefresh: true,
  });
}

export function fetchPasswordResetPreview(resetToken: string, signal?: AbortSignal): Promise<PasswordResetPreview> {
  return requestData<PasswordResetPreview>(`/auth/password-resets/${encodeURIComponent(resetToken)}`, {
    signal,
    skipSessionRefresh: true,
  });
}

export function resetPassword(resetToken: string, newPassword: string): Promise<PasswordResetResult> {
  return requestData<PasswordResetResult>("/auth/password-resets", {
    method: "POST",
    body: { token: resetToken, password: newPassword },
    skipSessionRefresh: true,
  });
}
