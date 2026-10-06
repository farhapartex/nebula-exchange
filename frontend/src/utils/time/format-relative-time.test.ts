import { describe, expect, it } from "vitest";

import { formatRelativeTime } from "@/utils/time/format-relative-time";

const now = new Date("2026-10-07T12:00:00Z");

describe("formatRelativeTime", () => {
  it("describes how far away a time is in days, hours or minutes", () => {
    expect(formatRelativeTime("2026-10-10T12:00:00Z", now)).toBe("in 3 days");
    expect(formatRelativeTime("2026-10-07T17:00:00Z", now)).toBe("in 5 hours");
    expect(formatRelativeTime("2026-10-07T11:30:00Z", now)).toBe("30 minutes ago");
    expect(formatRelativeTime("2026-10-07T12:00:20Z", now)).toBe("in less than a minute");
  });
});
