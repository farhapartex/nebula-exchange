"use client";

import { useDeferredValue, useState } from "react";

import type { MarketFilters } from "@/features/market/api/market-types";

const defaultMarketFilters: MarketFilters = { category: "ALL", rarity: "ALL", sort: "PRICE_LOW", search: "" };

export function useMarketFilters() {
  const [filters, setFilters] = useState<MarketFilters>(defaultMarketFilters);
  const deferredSearch = useDeferredValue(filters.search);
  const appliedFilters: MarketFilters = { ...filters, search: deferredSearch };

  function changeFilters(changes: Partial<MarketFilters>) {
    setFilters((currentFilters) => ({ ...currentFilters, ...changes }));
  }

  return { filters, appliedFilters, changeFilters };
}
