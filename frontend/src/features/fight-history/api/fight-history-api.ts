import { requestList } from "@/lib/api/api-client";

export type WonFightLevel = {
  id: string;
  number: number;
  title: string;
  chapter_number: number;
  chapter_title: string;
};

export type WonFight = {
  id: string;
  level: WonFightLevel;
  stars: number | null;
  duration_ms: number;
  damage_dealt: number;
  damage_taken: number;
  finished_at: string;
};

const wonFightsPageSize = 20;

export const wonFightsQueryKey = ["fight-sessions", "WON"] as const;

export function fetchWonFightsPage(cursor: string | null) {
  return requestList<WonFight>("/fight-sessions", { cursor, limit: wonFightsPageSize }, { query: { outcome: "WON" } });
}
