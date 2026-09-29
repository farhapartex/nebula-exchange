import type { AccountStatus } from "@/features/auth/api/auth-types";
import type { AuthStatus } from "@/features/auth/session/auth-context";

export const entryFeePath = "/onboarding/pay";

const publicGamePaths = ["/exchange", "/auctions"];
const pendingPaymentPaths = [...publicGamePaths, "/settings", entryFeePath];

export type GameRouteDecision = { kind: "allow" } | { kind: "wait" } | { kind: "redirect"; destination: string };

function matchesPathPrefix(currentPathname: string, allowedPaths: string[]): boolean {
  return allowedPaths.some(
    (allowedPath) => currentPathname === allowedPath || currentPathname.startsWith(`${allowedPath}/`),
  );
}

export function decideGameRouteAccess(
  currentPathname: string,
  authStatus: AuthStatus,
  accountStatus: AccountStatus | null,
): GameRouteDecision {
  if (matchesPathPrefix(currentPathname, publicGamePaths)) {
    return { kind: "allow" };
  }
  if (authStatus === "restoring") {
    return { kind: "wait" };
  }
  if (authStatus === "anonymous" || accountStatus === null) {
    return { kind: "redirect", destination: `/login?next=${encodeURIComponent(currentPathname)}` };
  }
  if (accountStatus === "PENDING_PAYMENT" && !matchesPathPrefix(currentPathname, pendingPaymentPaths)) {
    return { kind: "redirect", destination: entryFeePath };
  }
  if (accountStatus === "ACTIVE" && currentPathname === entryFeePath) {
    return { kind: "redirect", destination: "/hangar" };
  }
  return { kind: "allow" };
}
