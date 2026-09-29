"use client";

import { useQuery } from "@tanstack/react-query";

import { balancesQueryKey, fetchBalances } from "@/features/balances/api/balances-api";
import { useAuth } from "@/features/auth/session/use-auth";

const balanceRefreshIntervalInMilliseconds = 30_000;

export function useBalances() {
  const { status } = useAuth();
  return useQuery({
    queryKey: balancesQueryKey,
    queryFn: fetchBalances,
    enabled: status === "authenticated",
    refetchInterval: balanceRefreshIntervalInMilliseconds,
  });
}
