"use client";

import type { ComponentProps } from "react";
import { Tabs as TabsPrimitive } from "radix-ui";

import { cn } from "@/utils/class-names";

export const Tabs = TabsPrimitive.Root;

export function TabsList({ className, ...listProps }: ComponentProps<typeof TabsPrimitive.List>) {
  return (
    <TabsPrimitive.List
      className={cn(
        "inline-flex max-w-full items-center gap-1 overflow-x-auto rounded-lg border border-border bg-background/60 p-1",
        className,
      )}
      {...listProps}
    />
  );
}

export function TabsTrigger({ className, ...triggerProps }: ComponentProps<typeof TabsPrimitive.Trigger>) {
  return (
    <TabsPrimitive.Trigger
      className={cn(
        "inline-flex h-8 shrink-0 items-center rounded-md px-3 text-sm font-medium text-muted transition-colors",
        "hover:text-foreground focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none",
        "data-[state=active]:bg-surface-raised data-[state=active]:text-foreground data-[state=active]:shadow-sm",
        "disabled:pointer-events-none disabled:opacity-40",
        className,
      )}
      {...triggerProps}
    />
  );
}

export function TabsContent({ className, ...contentProps }: ComponentProps<typeof TabsPrimitive.Content>) {
  return <TabsPrimitive.Content className={cn("mt-4 focus-visible:outline-none", className)} {...contentProps} />;
}
