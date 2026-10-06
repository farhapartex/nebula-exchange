import { describe, expect, it } from "vitest";

import type { Plan, PlanOption } from "@/features/chapter-purchase/api/plan-api";
import { describePlanOption } from "@/features/chapter-purchase/plan-option-label";

function optionFor(chapterNumbers: number[]): PlanOption {
  return {
    chapter_count: chapterNumbers.length,
    chapters: chapterNumbers.map((chapterNumber) => ({
      id: String(chapterNumber),
      number: chapterNumber,
      title: `Title ${chapterNumber}`,
      price_cents: "499",
    })),
    subtotal_cents: "0",
    discount_percent: 0,
    discount_cents: "0",
    total_cents: "0",
  };
}

function planOfKind(kind: Plan["kind"]): Plan {
  return { id: kind, kind, name: kind, description: "", is_available: true, options: [] };
}

describe("describePlanOption", () => {
  it("names the chapters each plan option covers", () => {
    expect(describePlanOption(planOfKind("SINGLE_CHAPTER"), optionFor([2]))).toBe("Chapter 2 · Title 2");
    expect(describePlanOption(planOfKind("CHAPTER_BUNDLE"), optionFor([2, 3, 4]))).toBe("Chapters 2 to 4");
    expect(describePlanOption(planOfKind("ALL_CHAPTERS"), optionFor([2, 3, 4, 5]))).toBe("All 4 chapters out now");
  });
});
