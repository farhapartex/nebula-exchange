"use client";

import { Coins, Plus } from "lucide-react";

import { useCurrentPlayer } from "@/features/auth/session/use-current-player";
import { useCoinPackDialog } from "@/features/coin-packs/coin-pack-dialog-context";
import { formatCoinAmount } from "@/utils/money/format-coin-amount";

export function CoinBalanceButton() {
  const { openCoinPackDialog } = useCoinPackDialog();
  const coins = useCurrentPlayer().data?.coins;

  return (
    <button
      type="button"
      onClick={() => openCoinPackDialog()}
      aria-label={coins === undefined ? "Get coins" : `${formatCoinAmount(coins)} coins. Get more coins`}
      className="group flex h-9 items-center gap-2 rounded-lg border border-border-strong bg-surface-raised pr-1 pl-3 text-sm transition-colors hover:border-amber-400/60 focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none"
    >
      <Coins className="size-4 text-amber-400" aria-hidden="true" />
      <span className="font-mono text-foreground tabular-nums">
        {coins === undefined ? "—" : formatCoinAmount(coins)}
      </span>
      <span className="hidden text-xs text-subtle sm:inline">coins</span>
      <span className="flex size-7 items-center justify-center rounded-md bg-amber-400/15 text-amber-300 transition-colors group-hover:bg-amber-400/25">
        <Plus className="size-4" aria-hidden="true" />
      </span>
    </button>
  );
}
