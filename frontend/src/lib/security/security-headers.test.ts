import { describe, expect, it } from "vitest";

import { buildContentSecurityPolicy, buildSecurityHeaders } from "@/lib/security/security-headers";

const apiBaseUrl = "https://api.nebula.test/api/v1";

describe("buildContentSecurityPolicy", () => {
  it("allows the API origin and its WebSocket origin", () => {
    const policy = buildContentSecurityPolicy({ isDevelopment: false, apiBaseUrl });
    expect(policy).toContain("connect-src 'self' https://api.nebula.test wss://api.nebula.test");
    expect(policy).toContain("frame-ancestors 'none'");
    expect(policy).not.toContain("unsafe-eval");
  });

  it("relaxes script and connect sources only in development", () => {
    const policy = buildContentSecurityPolicy({ isDevelopment: true, apiBaseUrl: "http://localhost:8080/api/v1" });
    expect(policy).toContain("'unsafe-eval'");
    expect(policy).toContain("ws://localhost:*");
  });
});

describe("buildSecurityHeaders", () => {
  it("sends HSTS outside development only", () => {
    const productionHeaderNames = buildSecurityHeaders({ isDevelopment: false, apiBaseUrl }).map(
      (header) => header.key,
    );
    const developmentHeaderNames = buildSecurityHeaders({ isDevelopment: true, apiBaseUrl }).map(
      (header) => header.key,
    );
    expect(productionHeaderNames).toContain("Strict-Transport-Security");
    expect(developmentHeaderNames).not.toContain("Strict-Transport-Security");
  });
});
