import { describe, expect, it } from "vitest";

import {
  decideCurrentUserView,
  decideProtectedRouteAccess,
  loginPathForProtectedPage,
} from "@/features/auth/session/protected-route-access";
import { ApiError } from "@/lib/api/api-error";

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

  it("shows the page only after the profile from /me has loaded", () => {
    expect(decideCurrentUserView("pending", null)).toBe("loading");
    expect(decideCurrentUserView("success", null)).toBe("ready");
  });

  it("treats a 401 from /me as a rejected session and other failures as retryable", () => {
    expect(decideCurrentUserView("error", new ApiError(401, "UNAUTHORIZED", "Log in to continue"))).toBe(
      "session_rejected",
    );
    expect(decideCurrentUserView("error", new ApiError(500, "INTERNAL_ERROR", "Something went wrong"))).toBe("failed");
    expect(decideCurrentUserView("error", new Error("offline"))).toBe("failed");
  });
});
