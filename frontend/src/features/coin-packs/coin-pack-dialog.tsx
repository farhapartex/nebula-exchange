"use client";

import { useState } from "react";
import { usePathname } from "next/navigation";
import { useMutation, useQuery } from "@tanstack/react-query";
import { CreditCard } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { useCurrentPlayer } from "@/features/auth/session/use-current-player";
import { coinPacksQueryKey, createCoinPackCheckout, fetchCoinPacks } from "@/features/coin-packs/api/coin-pack-api";
import { CoinPackChoiceCard } from "@/features/coin-packs/coin-pack-choice-card";
import { cheapestCoinPack, extraCoinsPercent, smallestCoinPackCovering } from "@/features/coin-packs/coin-pack-value";
import { publishToastEvent } from "@/lib/notifications/toast-events";
import { formatCoinAmount } from "@/utils/money/format-coin-amount";
import { formatUsd } from "@/utils/money/format-usd";

type CoinPackDialogProps = {
  isOpen: boolean;
  missingCoins: bigint | null;
  onOpenChange: (isOpen: boolean) => void;
};

export function CoinPackDialog({ isOpen, missingCoins, onOpenChange }: CoinPackDialogProps) {
  const pathname = usePathname();
  const currentPlayerQuery = useCurrentPlayer();
  const coinPacksQuery = useQuery({ queryKey: coinPacksQueryKey, queryFn: fetchCoinPacks, enabled: isOpen });
  const coinPacks = coinPacksQuery.data ?? [];
  const [chosenCoinPackID, setChosenCoinPackID] = useState<string | null>(null);
  const checkoutMutation = useMutation({
    mutationFn: createCoinPackCheckout,
    onSuccess: (checkoutSession) => window.location.assign(checkoutSession.checkout_url),
    onError: () =>
      publishToastEvent({
        tone: "error",
        title: "Checkout could not start",
        description: "Nothing was charged. Try again in a moment.",
      }),
  });
  const isPaying = checkoutMutation.isPending || checkoutMutation.isSuccess;

  const suggestedCoinPack =
    missingCoins !== null && missingCoins > 0n
      ? smallestCoinPackCovering(coinPacks, missingCoins)
      : cheapestCoinPack(coinPacks);
  const selectedCoinPack = coinPacks.find((coinPack) => coinPack.id === chosenCoinPackID) ?? suggestedCoinPack;
  const baseCoinPack = cheapestCoinPack(coinPacks);
  const balanceCoins = BigInt(currentPlayerQuery.data?.coins ?? "0");

  function changeOpenState(nextIsOpen: boolean) {
    if (isPaying) {
      return;
    }
    if (!nextIsOpen) {
      setChosenCoinPackID(null);
    }
    onOpenChange(nextIsOpen);
  }

  function pay() {
    if (!selectedCoinPack) {
      return;
    }
    checkoutMutation.mutate({ coin_pack_id: selectedCoinPack.id, return_path: pathname });
  }

  return (
    <Dialog
      isOpen={isOpen}
      onOpenChange={changeOpenState}
      title="Get coins"
      description="Coins buy tools in the shop and from other players. They can't be turned back into money."
    >
      {coinPacksQuery.isError ? (
        <ErrorState
          message="The coin packs could not be loaded. Try again."
          onRetry={() => void coinPacksQuery.refetch()}
          isRetrying={coinPacksQuery.isFetching}
        />
      ) : !coinPacksQuery.data ? (
        <div className="space-y-3">
          <Skeleton className="h-12 rounded-xl" />
          <Skeleton className="h-20 rounded-xl" />
          <Skeleton className="h-20 rounded-xl" />
          <Skeleton className="h-20 rounded-xl" />
        </div>
      ) : !selectedCoinPack || !baseCoinPack ? (
        <p className="rounded-xl border border-border bg-background/60 p-4 text-sm text-muted">
          Coin packs are not on sale right now. Check back soon.
        </p>
      ) : (
        <div className="space-y-6">
          <div className="rounded-xl border border-border bg-background/60 px-4 py-3 text-sm">
            <dl className="flex items-center justify-between gap-2">
              <dt className="text-muted">Your coins</dt>
              <dd className="font-mono text-foreground tabular-nums">{formatCoinAmount(balanceCoins)}</dd>
            </dl>
            {missingCoins !== null && missingCoins > 0n && (
              <p className="mt-1 text-xs text-down">You need {formatCoinAmount(missingCoins)} more for that tool.</p>
            )}
          </div>

          <fieldset disabled={isPaying} className="space-y-3">
            <legend className="mb-2 text-xs font-semibold tracking-[0.18em] text-subtle uppercase">
              Choose a pack
            </legend>
            {coinPacks.map((coinPack) => (
              <CoinPackChoiceCard
                key={coinPack.id}
                coinPack={coinPack}
                extraPercent={extraCoinsPercent(coinPack, baseCoinPack)}
                isSelected={coinPack.id === selectedCoinPack.id}
                onSelect={setChosenCoinPackID}
              />
            ))}
          </fieldset>

          <p className="flex items-center gap-2 text-xs text-muted">
            <CreditCard className="size-4 shrink-0" aria-hidden="true" />
            You pay by card with Stripe. The coins are added as soon as the payment goes through.
          </p>

          <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
            <Button
              type="button"
              variant="secondary"
              size="lg"
              onClick={() => changeOpenState(false)}
              disabled={isPaying}
            >
              Cancel
            </Button>
            <Button type="button" size="lg" onClick={pay} isLoading={isPaying} className="sm:min-w-52">
              {isPaying ? "Opening Stripe" : `Pay ${formatUsd(BigInt(selectedCoinPack.price_cents))}`}
            </Button>
          </div>
        </div>
      )}
    </Dialog>
  );
}
