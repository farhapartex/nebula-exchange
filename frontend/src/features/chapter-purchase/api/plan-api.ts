import { requestAllPages } from "@/lib/api/api-client";

export type PlanKind = "SINGLE_CHAPTER" | "CHAPTER_BUNDLE" | "ALL_CHAPTERS";

export type PlanChapter = {
  id: string;
  number: number;
  title: string;
  price_cents: string;
};

export type PlanOption = {
  chapter_count: number;
  chapters: PlanChapter[];
  subtotal_cents: string;
  discount_percent: number;
  discount_cents: string;
  total_cents: string;
};

export type Plan = {
  id: string;
  kind: PlanKind;
  name: string;
  description: string;
  is_available: boolean;
  options: PlanOption[];
};

export const plansQueryKey = ["plans"] as const;

export function fetchPlans(): Promise<Plan[]> {
  return requestAllPages<Plan>("/plans");
}
