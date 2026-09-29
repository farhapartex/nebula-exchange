import { describe, expect, it } from "vitest";

import { formatRelativeTime } from "@/utils/time/relative-time";

const referenceNow = new Date("2026-09-29T12:00:00Z");

function secondsBefore(seconds: number) {
  return new Date(referenceNow.getTime() - seconds * 1000);
}

describe("formatRelativeTime", () => {
  it("shows just now for the last 45 seconds", () => {
    expect(formatRelativeTime(secondsBefore(10), referenceNow)).toBe("just now");
  });

  it("picks the largest fitting unit", () => {
    expect(formatRelativeTime(secondsBefore(120), referenceNow)).toBe("2 min. ago");
    expect(formatRelativeTime(secondsBefore(3 * 3600), referenceNow)).toBe("3 hr. ago");
    expect(formatRelativeTime(secondsBefore(86_400), referenceNow)).toBe("yesterday");
  });

  it("accepts ISO strings", () => {
    expect(formatRelativeTime("2026-09-29T11:00:00Z", referenceNow)).toBe("1 hr. ago");
  });
});
