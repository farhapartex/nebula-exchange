"use client";

import { useQuery } from "@tanstack/react-query";

import { fetchCurrentPlayer } from "@/features/auth/api/session-api";

export const currentPlayerQueryKey = ["current-player"] as const;

export function useVerifiedCurrentPlayer(isEnabled: boolean) {
  return useQuery({
    queryKey: currentPlayerQueryKey,
    queryFn: ({ signal }) => fetchCurrentPlayer(signal),
    enabled: isEnabled,
    staleTime: 0,
    refetchOnMount: "always",
  });
}

export function useCurrentPlayer() {
  return useQuery({
    queryKey: currentPlayerQueryKey,
    queryFn: ({ signal }) => fetchCurrentPlayer(signal),
    staleTime: Infinity,
  });
}
