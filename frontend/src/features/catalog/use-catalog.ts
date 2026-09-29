"use client";

import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";

import {
  fetchCatalogItems,
  fetchRecipes,
  fetchShopItems,
  fetchUpgrades,
  fetchZones,
} from "@/features/catalog/api/catalog-api";
import type { CatalogItem } from "@/features/catalog/api/catalog-types";

const catalogStaleTimeInMilliseconds = 5 * 60_000;

export const catalogQueryKeys = {
  items: ["catalog", "items"],
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
  return useQuery({ queryKey: catalogQueryKeys.zones, queryFn: fetchZones, staleTime: catalogStaleTimeInMilliseconds });
}

export function useShopItems() {
  return useQuery({
    queryKey: catalogQueryKeys.shopItems,
    queryFn: fetchShopItems,
    staleTime: catalogStaleTimeInMilliseconds,
  });
}
