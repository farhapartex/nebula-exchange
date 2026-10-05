import type { FighterProfile, LoadoutSlot, StoryChapter } from "@/features/fight-hub/api/fight-hub-types";
import { requestData } from "@/lib/api/api-client";

export const fightHubQueryKeys = {
  fighter: ["me", "fighter"] as const,
  story: ["me", "story"] as const,
  loadout: ["me", "loadout"] as const,
};

export function fetchFighterProfile(): Promise<FighterProfile> {
  return requestData<FighterProfile>("/me/fighter");
}

export function fetchStoryProgress(): Promise<StoryChapter[]> {
  return requestData<StoryChapter[]>("/me/story");
}

export function fetchLoadout(): Promise<LoadoutSlot[]> {
  return requestData<LoadoutSlot[]>("/me/loadout");
}
