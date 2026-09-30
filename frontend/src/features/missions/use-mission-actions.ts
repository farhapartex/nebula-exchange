"use client";

import { useQueryClient } from "@tanstack/react-query";

import { balancesQueryKey } from "@/features/balances/api/balances-api";
import { catalogQueryKeys } from "@/features/catalog/use-catalog";
import { onboardingProgressQueryKey } from "@/features/hangar/api/onboarding-api";
import { inventoryQueryKey } from "@/features/inventory/api/inventory-api";
import { missionsQueryKeys } from "@/features/missions/api/missions-api";

export function useRefreshAfterMissionChange() {
  const queryClient = useQueryClient();
  return () => {
    for (const queryKey of [
      missionsQueryKeys.all,
      inventoryQueryKey,
      catalogQueryKeys.zones,
      balancesQueryKey,
      onboardingProgressQueryKey,
    ]) {
      void queryClient.invalidateQueries({ queryKey });
    }
  };
}
