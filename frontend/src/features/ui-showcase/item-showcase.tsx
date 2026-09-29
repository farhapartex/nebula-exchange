"use client";

import { useState } from "react";

import { ContentSection } from "@/components/layout/content-section";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import type { ItemCategory } from "@/features/catalog/api/catalog-types";
import { ItemCard } from "@/features/catalog/components/item-card";
import { ItemIcon } from "@/features/catalog/components/item-icon";
import { useCatalogItems } from "@/features/catalog/use-catalog";

type CategoryFilter = ItemCategory | "all";

const categoryFilters: { value: CategoryFilter; label: string }[] = [
  { value: "all", label: "All" },
  { value: "resource", label: "Resources" },
  { value: "component", label: "Components" },
  { value: "drill", label: "Drills" },
  { value: "ship", label: "Ships" },
  { value: "consumable", label: "Consumables" },
  { value: "legendary", label: "Legendary" },
];

export function ItemShowcase() {
  const itemsQuery = useCatalogItems();
  const [categoryFilter, setCategoryFilter] = useState<CategoryFilter>("all");
  const visibleItems = (itemsQuery.data ?? []).filter(
    (catalogItem) => categoryFilter === "all" || catalogItem.category === categoryFilter,
  );

  return (
    <ContentSection title="ItemIcon and ItemCard" description="Every item in the catalog, loaded from GET /items.">
      {itemsQuery.isPending && (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 8 }, (_, skeletonIndex) => (
            <Skeleton key={skeletonIndex} className="h-36 rounded-2xl" />
          ))}
        </div>
      )}
      {itemsQuery.isError && (
        <ErrorState message="We couldn't load the item catalog." onRetry={() => void itemsQuery.refetch()} />
      )}
      {itemsQuery.isSuccess && (
        <div className="space-y-5">
          <div className="flex flex-wrap items-center gap-2">
            {itemsQuery.data.map((catalogItem) => (
              <ItemIcon key={catalogItem.id} item={catalogItem} size="sm" />
            ))}
          </div>
          <Tabs value={categoryFilter} onValueChange={(nextFilter) => setCategoryFilter(nextFilter as CategoryFilter)}>
            <TabsList>
              {categoryFilters.map((filterOption) => (
                <TabsTrigger key={filterOption.value} value={filterOption.value}>
                  {filterOption.label}
                </TabsTrigger>
              ))}
            </TabsList>
          </Tabs>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {visibleItems.map((catalogItem) => (
              <ItemCard key={catalogItem.id} item={catalogItem} />
            ))}
          </div>
        </div>
      )}
    </ContentSection>
  );
}
