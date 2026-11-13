import { Coins } from "lucide-react";

import type { CoinPack } from "@/features/coin-packs/api/coin-pack-api";
import { cn } from "@/utils/class-names";
import { formatCoinAmount } from "@/utils/money/format-coin-amount";
import { formatUsd } from "@/utils/money/format-usd";

type CoinPackChoiceCardProps = {
  coinPack: CoinPack;
  extraPercent: number;
  isSelected: boolean;
  onSelect: (coinPackID: string) => void;
};

export function CoinPackChoiceCard({ coinPack, extraPercent, isSelected, onSelect }: CoinPackChoiceCardProps) {
  return (
    <label
      className={cn(
        "flex cursor-pointer items-center gap-4 rounded-xl border p-4 transition-colors has-[:focus-visible]:ring-2 has-[:focus-visible]:ring-highlight",
        isSelected ? "border-accent bg-accent/10" : "border-border bg-background/40 hover:border-border-strong",
      )}
    >
      <input
        type="radio"
        name="coin-pack"
        value={coinPack.id}
        checked={isSelected}
        onChange={() => onSelect(coinPack.id)}
        className="sr-only"
      />
      <span
        className={cn(
          "flex size-11 shrink-0 items-center justify-center rounded-lg",
          isSelected ? "bg-amber-400/20 text-amber-300" : "bg-surface-raised text-amber-400",
        )}
      >
        <Coins className="size-5" aria-hidden="true" />
      </span>
      <span className="min-w-0 flex-1">
        <span className="flex flex-wrap items-center gap-2">
          <span className="font-mono text-lg text-foreground tabular-nums">{formatCoinAmount(coinPack.coins)}</span>
          <span className="text-sm text-muted">coins</span>
          {extraPercent > 0 && (
            <span className="rounded-md bg-up/15 px-2 py-0.5 text-xs font-semibold text-up">
              +{extraPercent}% extra
            </span>
          )}
        </span>
        <span className="mt-0.5 block text-xs text-subtle">{coinPack.name}</span>
      </span>
      <span className="font-mono text-sm whitespace-nowrap text-foreground tabular-nums">
        {formatUsd(BigInt(coinPack.price_cents))}
      </span>
    </label>
  );
}
