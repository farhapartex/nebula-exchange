"use client";

import Link from "next/link";

import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { ItemChip } from "@/features/catalog/components/item-chip";
import { useCatalogItems } from "@/features/catalog/use-catalog";
import { useInventory } from "@/features/inventory/use-inventory";

const fleetCategories = new Set(["ship", "drill", "consumable"]);

export function FleetCard() {
  const inventoryQuery = useInventory();
  const { itemsByID } = useCatalogItems();

  const fleetHoldings = (inventoryQuery.data ?? []).filter((holding) => {
    const catalogItem = itemsByID.get(holding.item_id);
    return catalogItem !== undefined && fleetCategories.has(catalogItem.category);
  });

  return (
    <Card>
      <CardHeader
        title="Your fleet"
        description="Ships, drills and fuel ready for missions."
        action={
          <Link href="/inventory" className="text-sm text-accent-soft hover:text-foreground">
            Inventory
          </Link>
        }
      />
      <CardContent>
        {inventoryQuery.isPending ? (
          <Skeleton className="h-20 rounded-xl" />
        ) : fleetHoldings.length === 0 ? (
          <p className="text-sm text-muted">
            No ships yet.{" "}
            <Link href="/shop" className="text-accent-soft hover:text-foreground">
              Visit the shop
            </Link>{" "}
            to get one.
          </p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {fleetHoldings.map((holding) => {
              const catalogItem = itemsByID.get(holding.item_id);
              return catalogItem ? (
                <ItemChip
                  key={holding.item_id}
                  item={catalogItem}
                  quantityLabel={String(holding.available + holding.held)}
                />
              ) : null;
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
