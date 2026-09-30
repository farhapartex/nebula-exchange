"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { PackageOpen } from "lucide-react";

import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import type { CatalogItem } from "@/features/catalog/api/catalog-types";
import { categoryFilters, matchesCategoryFilter, type CategoryFilter } from "@/features/catalog/category-filters";
import { CategoryFilterTabs } from "@/features/catalog/components/category-filter-tabs";
import { ItemCard } from "@/features/catalog/components/item-card";
import { useCatalogItems } from "@/features/catalog/use-catalog";
import type { InventoryHolding } from "@/features/inventory/api/inventory-api";
import { HoldingQuantity } from "@/features/inventory/holding-quantity";
import { useInventory } from "@/features/inventory/use-inventory";

type OwnedItem = { item: CatalogItem; holding: InventoryHolding };

export function InventoryView() {
  const inventoryQuery = useInventory();
  const catalogQuery = useCatalogItems();
  const [categoryFilter, setCategoryFilter] = useState<CategoryFilter>("all");

  const ownedItems = useMemo<OwnedItem[]>(() => {
    const ownedEntries: OwnedItem[] = [];
    for (const holding of inventoryQuery.data ?? []) {
      const catalogItem = catalogQuery.itemsByID.get(holding.item_id);
      if (catalogItem) {
        ownedEntries.push({ item: catalogItem, holding });
      }
    }
    return ownedEntries;
  }, [inventoryQuery.data, catalogQuery.itemsByID]);

  const countsByFilter = useMemo(() => {
    const counts: Partial<Record<CategoryFilter, number>> = {};
    for (const filterOption of categoryFilters) {
      counts[filterOption.value] = ownedItems.filter((ownedItem) =>
        matchesCategoryFilter(ownedItem.item.category, filterOption.value),
      ).length;
    }
    return counts;
  }, [ownedItems]);

  if (inventoryQuery.isPending || catalogQuery.isPending) {
    return (
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {Array.from({ length: 8 }, (_, skeletonIndex) => (
          <Skeleton key={skeletonIndex} className="h-44 rounded-2xl" />
        ))}
      </div>
    );
  }
  if (inventoryQuery.isError || catalogQuery.isError) {
    return (
      <ErrorState
        message="We couldn't load your inventory."
        onRetry={() => {
          void inventoryQuery.refetch();
          void catalogQuery.refetch();
        }}
      />
    );
  }
  if (ownedItems.length === 0) {
    return (
      <EmptyState
        icon={PackageOpen}
        title="Your hold is empty"
        description="Buy a ship and a drill in the shop, then send them on missions to bring back resources."
        action={
          <Button asChild>
            <Link href="/shop">Visit the shop</Link>
          </Button>
        }
      />
    );
  }

  const visibleItems = ownedItems.filter((ownedItem) => matchesCategoryFilter(ownedItem.item.category, categoryFilter));

  return (
    <div className="space-y-5">
      <CategoryFilterTabs value={categoryFilter} onValueChange={setCategoryFilter} countsByFilter={countsByFilter} />
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {visibleItems.map((ownedItem) => (
          <Link
            key={ownedItem.item.id}
            href={`/items/${ownedItem.item.id}`}
            className="rounded-2xl focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none"
          >
            <ItemCard
              item={ownedItem.item}
              className="h-full"
              footer={<HoldingQuantity holding={ownedItem.holding} />}
            />
          </Link>
        ))}
      </div>
    </div>
  );
}
