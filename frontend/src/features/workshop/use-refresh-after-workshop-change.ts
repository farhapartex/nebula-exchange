"use client";

import { useQueryClient } from "@tanstack/react-query";

import { balancesQueryKey } from "@/features/balances/api/balances-api";
import { onboardingProgressQueryKey } from "@/features/hangar/api/onboarding-api";
import { inventoryQueryKey } from "@/features/inventory/api/inventory-api";
import { workshopQueryKeys } from "@/features/workshop/api/workshop-api";

export function useRefreshAfterWorkshopChange() {
  const queryClient = useQueryClient();
  return () => {
    for (const queryKey of [
      workshopQueryKeys.crafts,
      inventoryQueryKey,
      balancesQueryKey,
      onboardingProgressQueryKey,
    ]) {
      void queryClient.invalidateQueries({ queryKey });
    }
  };
}
