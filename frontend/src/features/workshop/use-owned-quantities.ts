"use client";

import { useMemo } from "react";

import { useInventory } from "@/features/inventory/use-inventory";

export function useOwnedQuantities() {
  const inventoryQuery = useInventory();
  const ownedByItem = useMemo(
    () => new Map((inventoryQuery.data ?? []).map((holding) => [holding.item_id, holding.available])),
    [inventoryQuery.data],
  );
  return { ownedByItem, isPending: inventoryQuery.isPending };
}
