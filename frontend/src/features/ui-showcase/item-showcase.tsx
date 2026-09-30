"use client";

import { useState } from "react";

import { ContentSection } from "@/components/layout/content-section";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { matchesCategoryFilter, type CategoryFilter } from "@/features/catalog/category-filters";
import { CategoryFilterTabs } from "@/features/catalog/components/category-filter-tabs";
import { ItemCard } from "@/features/catalog/components/item-card";
import { ItemIcon } from "@/features/catalog/components/item-icon";
import { useCatalogItems } from "@/features/catalog/use-catalog";

export function ItemShowcase() {
  const itemsQuery = useCatalogItems();
  const [categoryFilter, setCategoryFilter] = useState<CategoryFilter>("all");
  const visibleItems = (itemsQuery.data ?? []).filter((catalogItem) =>
    matchesCategoryFilter(catalogItem.category, categoryFilter),
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
          <CategoryFilterTabs value={categoryFilter} onValueChange={setCategoryFilter} />
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
