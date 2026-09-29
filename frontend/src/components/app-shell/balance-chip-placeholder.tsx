import { Coins } from "lucide-react";

import { placeholderPilot } from "@/components/app-shell/placeholder-pilot";
import { NcAmount } from "@/components/money/nc-amount";

export function BalanceChipPlaceholder() {
  return (
    <div className="flex h-9 items-center gap-2 rounded-lg border border-border-strong bg-surface-raised px-3 text-sm text-foreground">
      <Coins className="size-4 text-warning" aria-hidden="true" />
      <NcAmount amount={placeholderPilot.totalBalanceInMicroUnits} />
    </div>
  );
}
