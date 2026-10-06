import { Coins } from "lucide-react";

import { formatCoinAmount } from "@/utils/money/format-coin-amount";

export function CoinPrice({ coins }: { coins: string }) {
  return (
    <span className="flex items-center gap-1.5 font-mono text-base font-semibold text-foreground tabular-nums">
      <Coins className="size-4 text-amber-400" aria-hidden="true" />
      {formatCoinAmount(coins)}
    </span>
  );
}
