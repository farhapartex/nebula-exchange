import { requestData } from "@/lib/api/api-client";

export type OnboardingStepKey = "paid_entry_fee" | "first_mission" | "first_craft" | "first_trade";

export type OnboardingProgress = {
  steps: { key: OnboardingStepKey; is_completed: boolean }[];
  completed_count: number;
};

export const onboardingProgressQueryKey = ["me", "onboarding"] as const;

export function fetchOnboardingProgress(): Promise<OnboardingProgress> {
  return requestData<OnboardingProgress>("/me/onboarding");
}
