import type { ItemQuantity } from "@/features/catalog/api/catalog-types";
import { requestData, requestList } from "@/lib/api/api-client";
import type { PaginationParameters } from "@/lib/api/api-types";

export type CraftStatus = "CRAFTING" | "DELIVERED";

export type CraftJob = {
  id: string;
  recipe_id: string;
  quantity: number;
  output_item_id: number;
  output_quantity: number;
  fee: string;
  inputs: ItemQuantity[];
  status: CraftStatus;
  started_at: string;
  ends_at: string;
  delivered_at: string | null;
};

export type CompletedUpgrade = {
  id: string;
  upgrade_id: string;
  path: "craft" | "buy";
  from_item_id: number;
  to_item_id: number;
  paid: string;
  journal_id: string;
};

export const workshopQueryKeys = {
  crafts: ["crafts"] as const,
  activeCraft: ["crafts", "active"] as const,
  craftHistory: ["crafts", "history"] as const,
};

export function startCraft(recipeID: string, quantity: number, idempotencyKey: string): Promise<CraftJob> {
  return requestData<CraftJob>("/crafts", { method: "POST", body: { recipe_id: recipeID, quantity }, idempotencyKey });
}

export function listCrafts(status: CraftStatus | null, pagination: PaginationParameters) {
  return requestList<CraftJob>("/crafts", pagination, { query: { status: status ?? undefined } });
}

export function craftUpgrade(upgradeID: string, idempotencyKey: string): Promise<CompletedUpgrade> {
  return requestData<CompletedUpgrade>(`/upgrades/${upgradeID}/crafts`, { method: "POST", idempotencyKey });
}

export function buyUpgrade(upgradeID: string, idempotencyKey: string): Promise<CompletedUpgrade> {
  return requestData<CompletedUpgrade>(`/upgrades/${upgradeID}/purchases`, { method: "POST", idempotencyKey });
}
