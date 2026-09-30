"use client";

import { Minus, Plus } from "lucide-react";

import { cn } from "@/utils/class-names";

type QuantityStepperProps = {
  value: number;
  onValueChange: (quantity: number) => void;
  minimum?: number;
  maximum: number;
  label: string;
  className?: string;
};

export function QuantityStepper({
  value,
  onValueChange,
  minimum = 1,
  maximum,
  label,
  className,
}: QuantityStepperProps) {
  const clampQuantity = (quantity: number) => Math.min(Math.max(Math.trunc(quantity) || minimum, minimum), maximum);
  const stepButtonClassName =
    "flex size-10 items-center justify-center text-muted transition-colors hover:bg-surface-raised hover:text-foreground focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none disabled:pointer-events-none disabled:opacity-40";

  return (
    <div
      className={cn(
        "inline-flex items-center overflow-hidden rounded-lg border border-border-strong bg-background/60",
        className,
      )}
    >
      <button
        type="button"
        aria-label={`Decrease ${label}`}
        className={stepButtonClassName}
        disabled={value <= minimum}
        onClick={() => onValueChange(clampQuantity(value - 1))}
      >
        <Minus className="size-4" />
      </button>
      <input
        aria-label={label}
        inputMode="numeric"
        value={value}
        onChange={(changeEvent) => onValueChange(clampQuantity(Number(changeEvent.target.value)))}
        className="h-10 w-14 border-x border-border bg-transparent text-center font-mono text-sm text-foreground tabular-nums focus:outline-none"
      />
      <button
        type="button"
        aria-label={`Increase ${label}`}
        className={stepButtonClassName}
        disabled={value >= maximum}
        onClick={() => onValueChange(clampQuantity(value + 1))}
      >
        <Plus className="size-4" />
      </button>
    </div>
  );
}
