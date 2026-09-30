"use client";

import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useBalances } from "@/features/balances/use-balances";
import { useCatalogItems, useRecipes, useUpgrades } from "@/features/catalog/use-catalog";
import { CraftQueue, useActiveCraft } from "@/features/workshop/craft-queue";
import { RecipeCard } from "@/features/workshop/recipe-card";
import { UpgradeCard } from "@/features/workshop/upgrade-card";
import { useOwnedQuantities } from "@/features/workshop/use-owned-quantities";

export function WorkshopView() {
  const { itemsByID } = useCatalogItems();
  const recipesQuery = useRecipes();
  const upgradesQuery = useUpgrades();
  const { ownedByItem } = useOwnedQuantities();
  const balancesQuery = useBalances();
  const activeCraftQuery = useActiveCraft();
  const availableNc = balancesQuery.data ? BigInt(balancesQuery.data.available) : null;

  return (
    <Tabs defaultValue="crafting">
      <TabsList>
        <TabsTrigger value="crafting">Crafting</TabsTrigger>
        <TabsTrigger value="upgrades">Upgrades</TabsTrigger>
      </TabsList>
      <TabsContent value="crafting" className="space-y-5">
        <CraftQueue itemsByID={itemsByID} />
        {recipesQuery.isPending && <Skeleton className="h-72 rounded-2xl" />}
        {recipesQuery.isError && (
          <ErrorState message="We couldn't load the recipes." onRetry={() => void recipesQuery.refetch()} />
        )}
        <div className="grid gap-4 lg:grid-cols-2">
          {recipesQuery.data?.map((recipe) => (
            <RecipeCard
              key={recipe.id}
              recipe={recipe}
              itemsByID={itemsByID}
              ownedByItem={ownedByItem}
              availableNc={availableNc}
              isWorkshopBusy={Boolean(activeCraftQuery.data)}
            />
          ))}
        </div>
      </TabsContent>
      <TabsContent value="upgrades">
        {upgradesQuery.isPending && <Skeleton className="h-72 rounded-2xl" />}
        {upgradesQuery.isError && (
          <ErrorState message="We couldn't load the upgrades." onRetry={() => void upgradesQuery.refetch()} />
        )}
        <div className="grid gap-4 lg:grid-cols-2">
          {upgradesQuery.data?.map((upgrade) => (
            <UpgradeCard
              key={upgrade.id}
              upgrade={upgrade}
              itemsByID={itemsByID}
              ownedByItem={ownedByItem}
              availableNc={availableNc}
            />
          ))}
        </div>
      </TabsContent>
    </Tabs>
  );
}
