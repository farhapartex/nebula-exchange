"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Clock } from "lucide-react";

import { NcAmount } from "@/components/money/nc-amount";
import { Button } from "@/components/ui/button";
import { QuantityStepper } from "@/components/ui/quantity-stepper";
import { useToast } from "@/components/ui/toast/use-toast";
import type { CatalogItem, Recipe } from "@/features/catalog/api/catalog-types";
import { ItemIcon } from "@/features/catalog/components/item-icon";
import { startCraft } from "@/features/workshop/api/workshop-api";
import { RequirementList } from "@/features/workshop/requirement-list";
import { hasAllRequirements, maximumAffordableBatches } from "@/features/workshop/requirement-rules";
import { useRefreshAfterWorkshopChange } from "@/features/workshop/use-refresh-after-workshop-change";
import { createIdempotencyKey } from "@/lib/api/api-client";
import { formatDurationShort } from "@/utils/time/format-duration";

const maximumBatchQuantity = 100;

type RecipeCardProps = {
  recipe: Recipe;
  itemsByID: Map<number, CatalogItem>;
  ownedByItem: Map<number, number>;
  availableNc: bigint | null;
  isWorkshopBusy: boolean;
};

export function RecipeCard({ recipe, itemsByID, ownedByItem, availableNc, isWorkshopBusy }: RecipeCardProps) {
  const { showToast } = useToast();
  const refreshAfterWorkshopChange = useRefreshAfterWorkshopChange();
  const [quantity, setQuantity] = useState(1);
  const outputItem = itemsByID.get(recipe.output_item_id);
  const totalFee = BigInt(recipe.fee) * BigInt(quantity);
  const hasInputs = hasAllRequirements(recipe.inputs, ownedByItem, quantity);
  const hasFee = availableNc === null || availableNc >= totalFee;
  const affordableBatches = Math.min(
    Math.max(maximumAffordableBatches(recipe.inputs, ownedByItem), 1),
    maximumBatchQuantity,
  );

  const craftMutation = useMutation({
    mutationFn: () => startCraft(recipe.id, quantity, createIdempotencyKey()),
    onSuccess: (craftJob) => {
      refreshAfterWorkshopChange();
      showToast({
        tone: "success",
        title: `Crafting ${craftJob.output_quantity} × ${outputItem?.name ?? "item"}`,
        description: `Ready in ${formatDurationShort(recipe.craft_seconds * quantity)}.`,
      });
    },
  });

  const blockedReason = isWorkshopBusy
    ? "Workshop busy"
    : !hasInputs
      ? "Not enough materials"
      : !hasFee
        ? "Not enough NC"
        : null;

  return (
    <article className="flex flex-col gap-4 rounded-2xl border border-border bg-surface/80 p-5">
      <div className="flex items-center gap-3">
        {outputItem && <ItemIcon item={outputItem} />}
        <div className="min-w-0 flex-1">
          <h3 className="text-base font-semibold text-foreground">
            {recipe.output_quantity} × {outputItem?.name}
          </h3>
          <p className="text-xs text-muted">{outputItem?.description}</p>
        </div>
      </div>
      <RequirementList
        requirements={recipe.inputs}
        ownedByItem={ownedByItem}
        itemsByID={itemsByID}
        multiplier={quantity}
      />
      <dl className="flex flex-wrap gap-x-5 gap-y-1 text-sm">
        <div className="flex items-center gap-1.5 text-muted">
          <Clock className="size-4" aria-hidden="true" />
          <dt className="sr-only">Time</dt>
          <dd className="text-foreground">{formatDurationShort(recipe.craft_seconds * quantity)}</dd>
        </div>
        <div className="flex items-center gap-1.5 text-muted">
          <dt>Fee</dt>
          <dd>
            <NcAmount amount={totalFee} className={hasFee ? "text-foreground" : "text-down"} />
          </dd>
        </div>
      </dl>
      {craftMutation.error && (
        <p role="alert" className="text-sm text-down">
          {craftMutation.error.message}
        </p>
      )}
      <div className="mt-auto flex flex-wrap items-center justify-between gap-3 border-t border-border pt-4">
        <QuantityStepper
          label="Batch size"
          value={quantity}
          maximum={maximumBatchQuantity}
          onValueChange={setQuantity}
        />
        <div className="flex items-center gap-2">
          {affordableBatches > 1 && quantity !== affordableBatches && (
            <Button variant="ghost" size="sm" onClick={() => setQuantity(affordableBatches)}>
              Max {affordableBatches}
            </Button>
          )}
          <Button
            isLoading={craftMutation.isPending}
            disabled={blockedReason !== null}
            onClick={() => craftMutation.mutate()}
          >
            {blockedReason ?? "Craft"}
          </Button>
        </div>
      </div>
    </article>
  );
}
