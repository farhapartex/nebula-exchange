import type { ReactNode } from "react";

import { cn } from "@/utils/class-names";

type PurchaseChoiceCardProps = {
  planID: string;
  isSelected: boolean;
  onSelect: (planID: string) => void;
  title: string;
  detail: string;
  price?: string;
  discountPercent?: number;
  children?: ReactNode;
};

export function PurchaseChoiceCard({
  planID,
  isSelected,
  onSelect,
  title,
  detail,
  price,
  discountPercent,
  children,
}: PurchaseChoiceCardProps) {
  return (
    <label
      className={cn(
        "block cursor-pointer rounded-xl border p-4 transition-colors",
        isSelected ? "border-accent bg-accent/10" : "border-border bg-background/40 hover:border-border-strong",
      )}
    >
      <div className="flex items-start gap-3">
        <input
          type="radio"
          name="chapter-purchase-plan"
          value={planID}
          checked={isSelected}
          onChange={() => onSelect(planID)}
          className="mt-1 size-4 accent-accent"
        />
        <div className="flex-1">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <span className="font-medium text-foreground">{title}</span>
            {price && <span className="font-mono text-sm whitespace-nowrap text-foreground tabular-nums">{price}</span>}
            {!!discountPercent && (
              <span className="rounded-md bg-up/15 px-2 py-0.5 text-xs font-semibold text-up">
                {discountPercent}% off
              </span>
            )}
          </div>
          <p className="mt-0.5 text-sm text-muted">{detail}</p>
          {children}
        </div>
      </div>
    </label>
  );
}
