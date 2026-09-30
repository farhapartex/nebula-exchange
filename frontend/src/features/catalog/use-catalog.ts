"use client";

import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";

import {
  fetchCatalogItem,
  fetchCatalogItems,
  fetchRecipes,
  fetchShopItems,
  fetchUpgrades,
  fetchZones,
} from "@/features/catalog/api/catalog-api";
import { useAuth } from "@/features/auth/session/use-auth";
import type { CatalogItem } from "@/features/catalog/api/catalog-types";

const catalogStaleTimeInMilliseconds = 5 * 60_000;

export const catalogQueryKeys = {
  items: ["catalog", "items"],
  item: (itemID: number) => ["catalog", "items", itemID] as const,
  recipes: ["catalog", "recipes"],
  upgrades: ["catalog", "upgrades"],
  zones: ["catalog", "zones"],
  shopItems: ["catalog", "shop-items"],
} as const;

export function useCatalogItems() {
  const itemsQuery = useQuery({
    queryKey: catalogQueryKeys.items,
    queryFn: fetchCatalogItems,
    staleTime: catalogStaleTimeInMilliseconds,
  });
  const itemsByID = useMemo(
    () => new Map<number, CatalogItem>((itemsQuery.data ?? []).map((catalogItem) => [catalogItem.id, catalogItem])),
    [itemsQuery.data],
  );
  return { ...itemsQuery, itemsByID };
}

export function useRecipes() {
  return useQuery({
    queryKey: catalogQueryKeys.recipes,
    queryFn: fetchRecipes,
    staleTime: catalogStaleTimeInMilliseconds,
  });
}

export function useUpgrades() {
  return useQuery({
    queryKey: catalogQueryKeys.upgrades,
    queryFn: fetchUpgrades,
    staleTime: catalogStaleTimeInMilliseconds,
  });
}

export function useZones() {
  const { user } = useAuth();
  return useQuery({
    queryKey: [...catalogQueryKeys.zones, user?.id ?? "guest"],
    queryFn: fetchZones,
    staleTime: user ? 0 : catalogStaleTimeInMilliseconds,
  });
}

export function useShopItems() {
  return useQuery({
    queryKey: catalogQueryKeys.shopItems,
    queryFn: fetchShopItems,
    staleTime: catalogStaleTimeInMilliseconds,
  });
}

export function useCatalogItem(itemID: number) {
  return useQuery({
    queryKey: catalogQueryKeys.item(itemID),
    queryFn: () => fetchCatalogItem(itemID),
    staleTime: catalogStaleTimeInMilliseconds,
    enabled: Number.isInteger(itemID) && itemID > 0,
  });
}
