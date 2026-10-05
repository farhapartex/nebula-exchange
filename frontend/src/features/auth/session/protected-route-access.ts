import type { AuthStatus } from "@/features/auth/session/auth-context";

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
