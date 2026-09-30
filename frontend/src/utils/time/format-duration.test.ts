import { describe, expect, it } from "vitest";

import {
  describeWaitTime,
  formatClockCountdown,
  formatCountdown,
  formatDurationShort,
} from "@/utils/time/format-duration";

describe("formatCountdown", () => {
  it("formats minutes and seconds", () => {
    expect(formatCountdown(840)).toBe("14:00");
    expect(formatCountdown(65.2)).toBe("1:06");
    expect(formatCountdown(0)).toBe("0:00");
    expect(formatCountdown(-5)).toBe("0:00");
  });
});

describe("describeWaitTime", () => {
  it("rounds up to whole minutes above one minute", () => {
    expect(describeWaitTime(840)).toBe("14 minutes");
    expect(describeWaitTime(61)).toBe("2 minutes");
    expect(describeWaitTime(60)).toBe("1 minute");
  });

  it("uses seconds under a minute", () => {
    expect(describeWaitTime(45)).toBe("45 seconds");
    expect(describeWaitTime(0.4)).toBe("1 second");
  });
});

describe("formatDurationShort", () => {
  it("uses the largest sensible units", () => {
    expect(formatDurationShort(45)).toBe("45s");
    expect(formatDurationShort(60)).toBe("1 min");
    expect(formatDurationShort(1800)).toBe("30 min");
    expect(formatDurationShort(3600)).toBe("1 h");
    expect(formatDurationShort(5400)).toBe("1 h 30 min");
    expect(formatDurationShort(14400)).toBe("4 h");
  });
});

describe("formatClockCountdown", () => {
  it("adds hours only when needed", () => {
    expect(formatClockCountdown(59)).toBe("0:59");
    expect(formatClockCountdown(900)).toBe("15:00");
    expect(formatClockCountdown(14_399)).toBe("3:59:59");
    expect(formatClockCountdown(-5)).toBe("0:00");
  });
});
