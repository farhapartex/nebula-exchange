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
