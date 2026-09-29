export type SessionEndReason = "signed_out" | "session_expired";

export const sessionEndReasonParameter = "reason";

export function loginPathAfterSessionEnd(reason: SessionEndReason, returnPath?: string): string {
  const searchParameters = new URLSearchParams({ [sessionEndReasonParameter]: reason });
  if (returnPath && returnPath !== "/login") {
    searchParameters.set("next", returnPath);
  }
  return `/login?${searchParameters.toString()}`;
}

export function parseSessionEndReason(rawReason: string | null): SessionEndReason | null {
  return rawReason === "signed_out" || rawReason === "session_expired" ? rawReason : null;
}
