"use client";

import { useQuery } from "@tanstack/react-query";

import { fetchCurrentUser } from "@/features/auth/api/session-api";

export const currentUserQueryKey = ["current-user"] as const;

export function useCurrentUser(isEnabled: boolean) {
  return useQuery({
    queryKey: currentUserQueryKey,
    queryFn: ({ signal }) => fetchCurrentUser(signal),
    enabled: isEnabled,
    staleTime: 0,
    refetchOnMount: "always",
  });
}
