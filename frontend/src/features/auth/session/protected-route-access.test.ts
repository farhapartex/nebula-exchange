import { describe, expect, it } from "vitest";

import { decideProtectedRouteAccess, loginPathForProtectedPage } from "@/features/auth/session/protected-route-access";

describe("protected route access", () => {
  it("renders only for an authenticated player", () => {
    expect(decideProtectedRouteAccess("authenticated", false)).toBe("render");
    expect(decideProtectedRouteAccess("authenticated", true)).toBe("render");
  });

  it("waits while the session is being restored", () => {
    expect(decideProtectedRouteAccess("restoring", false)).toBe("wait");
  });

  it("sends a visitor without a session to the login page", () => {
    expect(decideProtectedRouteAccess("anonymous", false)).toBe("redirect_to_login");
  });

  it("leaves navigation to the logout and expiry flows once a session has ended", () => {
    expect(decideProtectedRouteAccess("anonymous", true)).toBe("wait");
  });

  it("keeps the requested page so login can return to it", () => {
    expect(loginPathForProtectedPage("/fight/1-1")).toBe("/login?next=%2Ffight%2F1-1");
    expect(loginPathForProtectedPage("/fight?tab=story")).toBe("/login?next=%2Ffight%3Ftab%3Dstory");
    expect(loginPathForProtectedPage("/")).toBe("/login");
  });
});
