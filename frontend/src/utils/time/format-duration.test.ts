import { describe, expect, it } from "vitest";

import { describeWaitTime, formatCountdown } from "@/utils/time/format-duration";

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
