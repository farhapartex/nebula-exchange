"use client";

import { useQuery } from "@tanstack/react-query";

import { fetchInventory, inventoryQueryKey } from "@/features/inventory/api/inventory-api";

export function useInventory() {
  return useQuery({ queryKey: inventoryQueryKey, queryFn: fetchInventory });
}
