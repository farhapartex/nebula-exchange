"use client";

import Link from "next/link";
import { ArrowLeft, PackageSearch } from "lucide-react";

import { NcAmount } from "@/components/money/nc-amount";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { StatusBadge } from "@/components/ui/status-badge";
import { useAuth } from "@/features/auth/session/use-auth";
import type { CatalogItem, ItemQuantity } from "@/features/catalog/api/catalog-types";
import { ItemAttributeList } from "@/features/catalog/components/item-attributes";
import { ItemChip } from "@/features/catalog/components/item-chip";
import { ItemIcon } from "@/features/catalog/components/item-icon";
import { describeItemAppearance, describeItemTier } from "@/features/catalog/item-appearance";
import { formatLootChance, formatQuantityRange } from "@/features/catalog/loot-chance";
import {
  useCatalogItem,
  useCatalogItems,
  useRecipes,
  useShopItems,
  useUpgrades,
  useZones,
} from "@/features/catalog/use-catalog";
import { HoldingQuantity } from "@/features/inventory/holding-quantity";
import { useInventory } from "@/features/inventory/use-inventory";
import { UsageRow, UsageSection } from "@/features/item-detail/usage-section";
import { isApiError } from "@/lib/api/api-error";
import { formatDurationShort } from "@/utils/time/format-duration";

function ItemQuantityChips({
  quantities,
  itemsByID,
}: {
  quantities: ItemQuantity[];
  itemsByID: Map<number, CatalogItem>;
}) {
  return quantities.map((quantity) => {
    const quantityItem = itemsByID.get(quantity.item_id);
    return quantityItem ? (
      <ItemChip key={quantity.item_id} item={quantityItem} quantityLabel={String(quantity.quantity)} />
    ) : null;
  });
}

function OwnedQuantity({ itemID }: { itemID: number }) {
  const inventoryQuery = useInventory();
  const holding = inventoryQuery.data?.find((candidate) => candidate.item_id === itemID);
  if (!inventoryQuery.data) {
    return null;
  }
  return (
    <div className="mt-5 rounded-xl border border-border bg-background/50 px-4 py-3">
      {holding ? (
        <HoldingQuantity holding={holding} />
      ) : (
        <p className="text-sm text-muted">You don&apos;t own any yet.</p>
      )}
    </div>
  );
}

export function ItemDetailView({ itemID }: { itemID: number }) {
  const { status } = useAuth();
  const itemQuery = useCatalogItem(itemID);
  const catalogQuery = useCatalogItems();
  const recipesQuery = useRecipes();
  const upgradesQuery = useUpgrades();
  const zonesQuery = useZones();
  const shopItemsQuery = useShopItems();

  const isMissingItem =
    !Number.isInteger(itemID) ||
    itemID <= 0 ||
    (itemQuery.isError && isApiError(itemQuery.error) && itemQuery.error.statusCode === 404);
  if (isMissingItem) {
    return (
      <EmptyState
        icon={PackageSearch}
        title="This item does not exist"
        description="It may have been removed from the catalog."
        action={
          <Button asChild variant="secondary">
            <Link href="/inventory">Back to inventory</Link>
          </Button>
        }
      />
    );
  }
  if (itemQuery.isError) {
    return <ErrorState message="We couldn't load this item." onRetry={() => void itemQuery.refetch()} />;
  }
  if (!itemQuery.data || !catalogQuery.data) {
    return (
      <div className="grid gap-4 lg:grid-cols-[22rem_1fr]">
        <Skeleton className="h-80 rounded-2xl" />
        <div className="space-y-4">
          <Skeleton className="h-40 rounded-2xl" />
          <Skeleton className="h-40 rounded-2xl" />
        </div>
      </div>
    );
  }

  const itemDetail = itemQuery.data;
  const { itemsByID } = catalogQuery;
  const appearance = describeItemAppearance(itemDetail);
  const tierLabel = describeItemTier(itemDetail);
  const usage = itemDetail.usage;
  const recipes = recipesQuery.data ?? [];
  const upgrades = upgradesQuery.data ?? [];
  const zones = zonesQuery.data ?? [];
  const shopItems = shopItemsQuery.data ?? [];

  const craftingRows = recipes
    .filter((recipe) => usage.crafted_by.includes(recipe.id))
    .map((recipe) => (
      <UsageRow
        key={recipe.id}
        label="Craft in the workshop"
        detail={
          <>
            {formatDurationShort(recipe.craft_seconds)} · fee <NcAmount amount={recipe.fee} />
          </>
        }
      >
        <ItemQuantityChips quantities={recipe.inputs} itemsByID={itemsByID} />
      </UsageRow>
    ));
  const upgradeSourceRows = upgrades
    .filter((upgrade) => usage.upgraded_from.includes(upgrade.id))
    .map((upgrade) => (
      <UsageRow
        key={upgrade.id}
        label={`Upgrade from ${itemsByID.get(upgrade.from_item_id)?.name ?? "an earlier tier"}`}
        detail={
          <>
            fee <NcAmount amount={upgrade.craft_fee} />
            {upgrade.buy_price && (
              <>
                {" "}
                · or buy for <NcAmount amount={upgrade.buy_price} />
              </>
            )}
          </>
        }
      >
        <ItemQuantityChips quantities={upgrade.inputs} itemsByID={itemsByID} />
      </UsageRow>
    ));
  const lootRows = zones.flatMap((zone) =>
    zone.loot
      .filter((lootEntry) => lootEntry.item_id === itemDetail.id)
      .map((lootEntry) => (
        <UsageRow
          key={zone.id}
          label={`Mine in ${zone.name}`}
          detail={`${formatQuantityRange(lootEntry.minimum_quantity, lootEntry.maximum_quantity)} per run · ${formatLootChance(lootEntry.chance_basis_points)}`}
        />
      )),
  );
  const shopRows = shopItems
    .filter((shopItem) => usage.sold_as_skus.includes(shopItem.sku))
    .map((shopItem) => (
      <UsageRow key={shopItem.sku} label={`Shop: ${shopItem.name}`} detail={<NcAmount amount={shopItem.price} />}>
        <ItemQuantityChips quantities={shopItem.contents} itemsByID={itemsByID} />
      </UsageRow>
    ));

  const recipeOutputItems = recipes
    .filter((recipe) => usage.input_to_recipes.includes(recipe.id))
    .map((recipe) => itemsByID.get(recipe.output_item_id));
  const upgradeMaterialTargets = upgrades
    .filter((upgrade) => usage.input_to_upgrades.includes(upgrade.id))
    .map((upgrade) => itemsByID.get(upgrade.to_item_id));
  const upgradeTargets = upgrades
    .filter((upgrade) => usage.upgrades_into.includes(upgrade.id))
    .map((upgrade) => itemsByID.get(upgrade.to_item_id));
  const usedForGroups = [
    { label: "Crafting ingredient for", items: recipeOutputItems },
    { label: "Upgrade material for", items: upgradeMaterialTargets },
    { label: "Upgrades into", items: upgradeTargets },
  ];
  const usedForRows = usedForGroups
    .filter((usedForGroup) => usedForGroup.items.length > 0)
    .map((usedForGroup) => (
      <UsageRow key={usedForGroup.label} label={usedForGroup.label}>
        {usedForGroup.items.map((targetItem) => targetItem && <ItemChip key={targetItem.id} item={targetItem} />)}
      </UsageRow>
    ));

  return (
    <div className="space-y-6">
      {status === "authenticated" && (
        <Link href="/inventory" className="inline-flex items-center gap-1.5 text-sm text-muted hover:text-foreground">
          <ArrowLeft className="size-4" aria-hidden="true" />
          Inventory
        </Link>
      )}
      <div className="grid gap-4 lg:grid-cols-[22rem_1fr]">
        <aside className="h-fit rounded-2xl border border-border bg-surface/80 p-5">
          <div className="flex flex-col items-center text-center">
            <ItemIcon item={itemDetail} size="lg" />
            <h1 className="mt-4 text-2xl font-semibold tracking-tight text-foreground">{itemDetail.name}</h1>
            <p className="mt-1 text-sm text-muted">
              {appearance.label}
              {tierLabel && ` · ${tierLabel}`}
            </p>
            <div className="mt-3 flex flex-wrap justify-center gap-1.5">
              {itemDetail.is_tradeable && <StatusBadge label="Tradeable" tone="info" />}
              {itemDetail.is_auction_only && <StatusBadge label="Auction only" tone="warning" />}
              {itemDetail.max_supply !== null && (
                <StatusBadge label={`Max supply ${itemDetail.max_supply}`} tone="accent" />
              )}
            </div>
          </div>
          <p className="mt-5 text-sm leading-relaxed text-muted">{itemDetail.description}</p>
          <div className="mt-4">
            <ItemAttributeList item={itemDetail} />
          </div>
          {status === "authenticated" && <OwnedQuantity itemID={itemDetail.id} />}
        </aside>
        <div className="space-y-4">
          <UsageSection title="How to get it" emptyMessage="Only available from auctions and other players.">
            {[...craftingRows, ...upgradeSourceRows, ...lootRows, ...shopRows]}
          </UsageSection>
          <UsageSection title="Used for" emptyMessage="Nothing consumes this item. It is used as it is.">
            {usedForRows}
          </UsageSection>
        </div>
      </div>
    </div>
  );
}
