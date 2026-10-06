import type { Plan, PlanOption } from "@/features/chapter-purchase/api/plan-api";

export function describePlanOption(plan: Plan, option: PlanOption): string {
  const firstChapter = option.chapters[0];
  const lastChapter = option.chapters[option.chapters.length - 1];
  if (plan.kind === "SINGLE_CHAPTER") {
    return `Chapter ${firstChapter.number} · ${firstChapter.title}`;
  }
  if (plan.kind === "ALL_CHAPTERS") {
    return `All ${option.chapter_count} chapters out now`;
  }
  return `Chapters ${firstChapter.number} to ${lastChapter.number}`;
}
