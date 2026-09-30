import { requestAllPages } from "@/lib/api/api-client";

export type InventoryHolding = {
  item_id: number;
  available: number;
  held: number;
};

export const inventoryQueryKey = ["me", "inventory"] as const;

export function fetchInventory(): Promise<InventoryHolding[]> {
  return requestAllPages<InventoryHolding>("/me/inventory");
}
