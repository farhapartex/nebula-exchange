"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";

import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { buyMarketListing, buyShopItem, marketQueryKeys } from "@/features/market/api/market-api";
import type { MarketListing, ShopItem } from "@/features/market/api/market-types";
import { MasteryMeter } from "@/features/market/mastery-meter";
import { RarityBadge } from "@/features/market/rarity-badge";
import { ToolArtwork } from "@/features/market/tool-artwork";
import { fightHubQueryKeys } from "@/features/fight-hub/api/fight-hub-api";
import { useFighterProfile } from "@/features/fight-hub/use-fight-hub";
import { isApiError } from "@/lib/api/api-error";
import { publishToastEvent } from "@/lib/notifications/toast-events";
import { formatCoinAmount } from "@/utils/money/format-coin-amount";

export type PurchaseTarget = { kind: "SHOP"; shopItem: ShopItem } | { kind: "LISTING"; listing: MarketListing };

type PurchaseToolDialogProps = {
  purchaseTarget: PurchaseTarget | null;
  onClose: () => void;
};

function purchaseDetails(purchaseTarget: PurchaseTarget) {
  if (purchaseTarget.kind === "SHOP") {
    return {
      toolType: purchaseTarget.shopItem.tool_type,
      priceCoins: purchaseTarget.shopItem.price_coins,
      listing: null,
    };
  }
  return {
    toolType: purchaseTarget.listing.tool.tool_type,
    priceCoins: purchaseTarget.listing.price_coins,
    listing: purchaseTarget.listing,
  };
}

export function PurchaseToolDialog({ purchaseTarget, onClose }: PurchaseToolDialogProps) {
  const queryClient = useQueryClient();
  const fighterQuery = useFighterProfile();
  const purchaseMutation = useMutation({
    mutationFn: (target: PurchaseTarget) =>
      target.kind === "SHOP" ? buyShopItem(target.shopItem.tool_type.id) : buyMarketListing(target.listing.id),
    onSuccess: (_, target) => {
      const toolName = purchaseDetails(target).toolType.name;
      publishToastEvent({ tone: "success", title: `${toolName} is yours`, description: "Find it in your Arsenal." });
      void queryClient.invalidateQueries({ queryKey: fightHubQueryKeys.fighter });
      void queryClient.invalidateQueries({ queryKey: marketQueryKeys.allShopItems });
      void queryClient.invalidateQueries({ queryKey: marketQueryKeys.allListings });
      onClose();
    },
    onError: (error) =>
      publishToastEvent({
        tone: "error",
        title: "Purchase failed",
        description: isApiError(error) ? error.message : "Nothing was taken. Try again in a moment.",
      }),
  });

  if (!purchaseTarget) {
    return null;
  }
  const { toolType, priceCoins, listing } = purchaseDetails(purchaseTarget);
  const balanceCoins = BigInt(fighterQuery.data?.coins ?? 0);
  const price = BigInt(priceCoins);
  const missingCoins = price - balanceCoins;
  const hasEnoughCoins = missingCoins <= 0n;

  return (
    <Dialog
      isOpen
      onOpenChange={(isOpen) => !isOpen && !purchaseMutation.isPending && onClose()}
      title={`Buy ${toolType.name}`}
      description={
        listing
          ? `From ${listing.seller.username}. The copy keeps its mastery.`
          : "A new copy from the shop, at mastery 1."
      }
    >
      <div className="space-y-5">
        <div className="flex gap-4">
          <ToolArtwork toolType={toolType} className="w-32 shrink-0" />
          <div className="min-w-0 flex-1 space-y-2">
            <RarityBadge rarity={toolType.rarity} />
            <p className="text-sm text-muted">{toolType.description}</p>
            {listing && (
              <MasteryMeter masteryLevel={listing.tool.mastery_level} maxMasteryLevel={toolType.max_mastery_level} />
            )}
          </div>
        </div>
        {purchaseTarget.kind === "SHOP" && !purchaseTarget.shopItem.is_usable && (
          <p className="rounded-lg border border-amber-400/30 bg-amber-400/10 px-3 py-2 text-sm text-amber-200">
            You can buy it now and keep it in your Arsenal, but you can only use it from fighter level{" "}
            {toolType.minimum_fighter_level}.
          </p>
        )}
        <dl className="space-y-1.5 rounded-xl border border-border bg-background/60 p-4 text-sm">
          <div className="flex justify-between">
            <dt className="text-muted">Price</dt>
            <dd className="font-mono text-foreground tabular-nums">{formatCoinAmount(price)} coins</dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-muted">Your coins</dt>
            <dd className="font-mono text-foreground tabular-nums">{formatCoinAmount(balanceCoins)}</dd>
          </div>
          <div className="flex justify-between border-t border-border pt-1.5">
            <dt className="text-muted">After buying</dt>
            <dd
              className={hasEnoughCoins ? "font-mono text-foreground tabular-nums" : "font-mono text-down tabular-nums"}
            >
              {hasEnoughCoins ? formatCoinAmount(balanceCoins - price) : `${formatCoinAmount(missingCoins)} short`}
            </dd>
          </div>
        </dl>
        <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
          <Button type="button" variant="secondary" onClick={onClose} disabled={purchaseMutation.isPending}>
            Cancel
          </Button>
          <Button
            type="button"
            onClick={() => purchaseMutation.mutate(purchaseTarget)}
            isLoading={purchaseMutation.isPending}
            disabled={!hasEnoughCoins}
            className="sm:min-w-44"
          >
            {hasEnoughCoins
              ? `Buy for ${formatCoinAmount(price)} coins`
              : `Need ${formatCoinAmount(missingCoins)} more coins`}
          </Button>
        </div>
      </div>
    </Dialog>
  );
}
