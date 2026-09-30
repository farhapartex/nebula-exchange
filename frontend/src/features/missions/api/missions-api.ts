import { requestData, requestList } from "@/lib/api/api-client";
import type { PaginationParameters } from "@/lib/api/api-types";

export type MissionStatus = "RUNNING" | "COMPLETED" | "COLLECTED" | "ABORTED";

export type ExpectedLoot = {
  item_id: number;
  minimum_quantity: number;
  maximum_quantity: number;
  chance_basis_points: number;
};

export type LootRoll = {
  item_id: number;
  quantity: number;
};

export type Mission = {
  id: string;
  zone_id: string;
  zone_name: string;
  ship_item_id: number;
  drill_item_id: number;
  status: MissionStatus;
  fuel_spent: number;
  cargo_capacity: number;
  expected_loot: ExpectedLoot[];
  loot: LootRoll[];
  started_at: string;
  ends_at: string;
  resolved_at: string | null;
  collected_at: string | null;
  aborted_at: string | null;
};

export type StartMissionRequest = {
  zone_id: string;
  ship_item_id: number;
  drill_item_id: number;
};

export const missionsQueryKeys = {
  all: ["missions"] as const,
  active: ["missions", "active"] as const,
  history: ["missions", "history"] as const,
};

export function startMission(startRequest: StartMissionRequest, idempotencyKey: string): Promise<Mission> {
  return requestData<Mission>("/missions", { method: "POST", body: startRequest, idempotencyKey });
}

export function collectMission(missionID: string): Promise<Mission> {
  return requestData<Mission>(`/missions/${missionID}/collect`, { method: "POST" });
}

export function abortMission(missionID: string): Promise<Mission> {
  return requestData<Mission>(`/missions/${missionID}/abort`, { method: "POST" });
}

export function listMissions(statuses: MissionStatus[], pagination: PaginationParameters) {
  return requestList<Mission>("/missions", pagination, { query: { status: statuses.join(",") } });
}
