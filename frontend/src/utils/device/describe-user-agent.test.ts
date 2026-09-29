import { describe, expect, it } from "vitest";

import { describeUserAgent } from "@/utils/device/describe-user-agent";

describe("describeUserAgent", () => {
  it("names common browsers and systems", () => {
    expect(
      describeUserAgent(
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0 Safari/537.36",
      ),
    ).toBe("Chrome on macOS");
    expect(
      describeUserAgent(
        "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 Version/18.0 Mobile Safari/604.1",
      ),
    ).toBe("Safari on iOS");
    expect(describeUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0")).toBe(
      "Firefox on Windows",
    );
    expect(describeUserAgent("Mozilla/5.0 (Windows NT 10.0) Chrome/131.0 Safari/537.36 Edg/131.0")).toBe(
      "Edge on Windows",
    );
  });

  it("falls back for empty or unknown agents", () => {
    expect(describeUserAgent("")).toBe("Unknown device");
    expect(describeUserAgent("curl/8.7.1")).toBe("Unknown device");
  });
});
