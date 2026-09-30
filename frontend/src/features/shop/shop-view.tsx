"use client";

import { useState } from "react";

import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import type { ShopItem } from "@/features/catalog/api/catalog-types";
import { useCatalogItems, useShopItems } from "@/features/catalog/use-catalog";
import { PurchaseModal } from "@/features/shop/purchase-modal";
import { ShopItemCard } from "@/features/shop/shop-item-card";

export function ShopView() {
  const shopItemsQuery = useShopItems();
  const { itemsByID } = useCatalogItems();
  const [shopItemBeingBought, setShopItemBeingBought] = useState<ShopItem | null>(null);

  if (shopItemsQuery.isPending) {
    return (
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {Array.from({ length: 4 }, (_, skeletonIndex) => (
          <Skeleton key={skeletonIndex} className="h-56 rounded-2xl" />
        ))}
      </div>
    );
  }
  if (shopItemsQuery.isError) {
    return <ErrorState message="We couldn't load the shop." onRetry={() => void shopItemsQuery.refetch()} />;
  }

  return (
    <>
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {shopItemsQuery.data.map((shopItem) => (
          <ShopItemCard key={shopItem.sku} shopItem={shopItem} itemsByID={itemsByID} onBuy={setShopItemBeingBought} />
        ))}
      </div>
      <PurchaseModal
        shopItem={shopItemBeingBought}
        itemsByID={itemsByID}
        onClose={() => setShopItemBeingBought(null)}
      />
    </>
  );
}
