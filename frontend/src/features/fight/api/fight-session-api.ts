import { requestData } from "@/lib/api/api-client";

export type StartedFight = {
  id: string;
  level: string;
  status: "STARTED";
  seed: string;
  started_at: string;
};

export function startFightSession(levelID: string): Promise<StartedFight> {
  return requestData<StartedFight>("/fight-sessions", { method: "POST", body: { level: levelID } });
}

export type FightResultReport = {
  outcome: "WON" | "LOST";
  duration_ms: number;
  damage_dealt: number;
  damage_taken: number;
};

export type FightResult = {
  id: string;
  level: string;
  status: "FINISHED";
  outcome: "WON" | "LOST";
  stars: number;
  duration_ms: number;
  damage_dealt: number;
  damage_taken: number;
  is_level_completed: boolean;
  reward_coins: string;
  reward_experience: number;
  finished_at: string;
};

export function submitFightResult(fightSessionID: string, report: FightResultReport): Promise<FightResult> {
  return requestData<FightResult>(`/fight-sessions/${encodeURIComponent(fightSessionID)}/results`, {
    method: "POST",
    body: report,
  });
}
