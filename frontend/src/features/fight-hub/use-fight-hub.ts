"use client";

import { useQuery } from "@tanstack/react-query";

import {
  fetchFighterProfile,
  fetchLoadout,
  fetchStoryProgress,
  fightHubQueryKeys,
} from "@/features/fight-hub/api/fight-hub-api";
import { fetchNextLevel, nextLevelQueryKey } from "@/features/fight-hub/api/next-level-api";

export function useFighterProfile() {
  return useQuery({ queryKey: fightHubQueryKeys.fighter, queryFn: fetchFighterProfile });
}

export function useStoryProgress() {
  return useQuery({ queryKey: fightHubQueryKeys.story, queryFn: fetchStoryProgress });
}

export function useLoadout() {
  return useQuery({ queryKey: fightHubQueryKeys.loadout, queryFn: fetchLoadout });
}

export function useNextLevel() {
  return useQuery({
    queryKey: nextLevelQueryKey,
    queryFn: ({ signal }) => fetchNextLevel(signal),
    staleTime: 0,
    refetchOnMount: "always",
  });
}
