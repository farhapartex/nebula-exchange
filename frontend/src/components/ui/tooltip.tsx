"use client";

import type { ReactNode } from "react";
import { Tooltip as TooltipPrimitive } from "radix-ui";

export const TooltipProvider = TooltipPrimitive.Provider;

type TooltipProps = {
  content: ReactNode;
  children: ReactNode;
  side?: "top" | "right" | "bottom" | "left";
};

export function Tooltip({ content, children, side = "top" }: TooltipProps) {
  return (
    <TooltipPrimitive.Root>
      <TooltipPrimitive.Trigger asChild>{children}</TooltipPrimitive.Trigger>
      <TooltipPrimitive.Portal>
        <TooltipPrimitive.Content
          side={side}
          sideOffset={6}
          className="z-50 rounded-md border border-border-strong bg-surface-raised px-2.5 py-1.5 text-xs text-foreground shadow-lg shadow-black/50 data-[state=delayed-open]:animate-fade-in"
        >
          {content}
          <TooltipPrimitive.Arrow className="fill-surface-raised" />
        </TooltipPrimitive.Content>
      </TooltipPrimitive.Portal>
    </TooltipPrimitive.Root>
  );
}
