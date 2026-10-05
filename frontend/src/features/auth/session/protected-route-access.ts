import type { AuthStatus } from "@/features/auth/session/auth-context";
import { isApiError } from "@/lib/api/api-error";

export type ProtectedRouteAccess = "render" | "wait" | "redirect_to_login";

export function decideProtectedRouteAccess(status: AuthStatus, hasBeenAuthenticated: boolean): ProtectedRouteAccess {
  if (status === "authenticated") {
    return "render";
  }
  if (status === "anonymous" && !hasBeenAuthenticated) {
    return "redirect_to_login";
  }
  return "wait";
}

export function loginPathForProtectedPage(returnPath: string): string {
  if (!returnPath || returnPath === "/" || returnPath.startsWith("/login")) {
    return "/login";
  }
  return `/login?${new URLSearchParams({ next: returnPath }).toString()}`;
}

export type CurrentUserView = "loading" | "ready" | "session_rejected" | "failed";

export function decideCurrentUserView(queryStatus: "pending" | "error" | "success", error: unknown): CurrentUserView {
  if (queryStatus === "success") {
    return "ready";
  }
  if (queryStatus === "pending") {
    return "loading";
  }
  return isApiError(error) && error.statusCode === 401 ? "session_rejected" : "failed";
}
