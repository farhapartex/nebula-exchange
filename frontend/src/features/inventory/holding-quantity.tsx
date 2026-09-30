import { Lock } from "lucide-react";

import type { InventoryHolding } from "@/features/inventory/api/inventory-api";

export function HoldingQuantity({ holding }: { holding: InventoryHolding }) {
  return (
    <div className="flex items-center justify-between gap-3 text-sm">
      <span className="text-muted">
        Owned{" "}
        <span className="font-mono font-semibold text-foreground tabular-nums">{holding.available + holding.held}</span>
      </span>
      {holding.held > 0 ? (
        <span className="flex items-center gap-1 text-xs text-warning">
          <Lock className="size-3" aria-hidden="true" />
          <span className="font-mono tabular-nums">{holding.held}</span> held
        </span>
      ) : (
        <span className="text-xs text-subtle">All available</span>
      )}
    </div>
  );
}
