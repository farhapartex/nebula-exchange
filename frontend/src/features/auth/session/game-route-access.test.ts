import { describe, expect, it } from "vitest";

import { decideGameRouteAccess } from "@/features/auth/session/game-route-access";

describe("decideGameRouteAccess", () => {
  it("lets anyone view public markets, auctions and item pages", () => {
    expect(decideGameRouteAccess("/exchange", "anonymous", null)).toEqual({ kind: "allow" });
    expect(decideGameRouteAccess("/auctions/42", "restoring", null)).toEqual({ kind: "allow" });
    expect(decideGameRouteAccess("/items/301", "anonymous", null)).toEqual({ kind: "allow" });
    expect(decideGameRouteAccess("/items/301", "authenticated", "PENDING_PAYMENT")).toEqual({ kind: "allow" });
  });

  it("waits while the session is being restored", () => {
    expect(decideGameRouteAccess("/hangar", "restoring", null)).toEqual({ kind: "wait" });
  });

  it("sends guests to login with a return path", () => {
    expect(decideGameRouteAccess("/missions", "anonymous", null)).toEqual({
      kind: "redirect",
      destination: "/login?next=%2Fmissions",
    });
  });

  it("sends pending players to the entry fee except for allowed pages", () => {
    expect(decideGameRouteAccess("/hangar", "authenticated", "PENDING_PAYMENT")).toEqual({
      kind: "redirect",
      destination: "/onboarding/pay",
    });
    expect(decideGameRouteAccess("/settings", "authenticated", "PENDING_PAYMENT")).toEqual({ kind: "allow" });
    expect(decideGameRouteAccess("/payment/result", "authenticated", "PENDING_PAYMENT")).toEqual({ kind: "allow" });
    expect(decideGameRouteAccess("/onboarding/pay", "authenticated", "PENDING_PAYMENT")).toEqual({ kind: "allow" });
  });

  it("lets active and frozen players in, and moves active players past the entry fee", () => {
    expect(decideGameRouteAccess("/hangar", "authenticated", "ACTIVE")).toEqual({ kind: "allow" });
    expect(decideGameRouteAccess("/wallet", "authenticated", "FROZEN")).toEqual({ kind: "allow" });
    expect(decideGameRouteAccess("/onboarding/pay", "authenticated", "ACTIVE")).toEqual({
      kind: "redirect",
      destination: "/hangar",
    });
  });
});
