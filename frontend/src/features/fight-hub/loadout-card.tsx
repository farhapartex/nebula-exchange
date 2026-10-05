"use client";

import { Plus } from "lucide-react";

import { Skeleton } from "@/components/ui/skeleton";
import { HubPanel } from "@/features/fight-hub/hub-panel";
import { toolIcons } from "@/features/fight-hub/tool-icons";
import { useLoadout } from "@/features/fight-hub/use-fight-hub";

export function LoadoutCard() {
  const loadoutQuery = useLoadout();

  return (
    <HubPanel title="Your loadout" action={<span className="text-xs text-subtle">Change loadout · soon</span>}>
      {!loadoutQuery.data ? (
        <Skeleton className="h-24 rounded-xl" />
      ) : (
        <ul className="grid gap-3 sm:grid-cols-3">
          {loadoutQuery.data.map((loadoutSlot) => {
            if (!loadoutSlot.tool) {
              return (
                <li
                  key={loadoutSlot.slot_number}
                  className="flex min-h-28 flex-col items-center justify-center gap-1.5 rounded-xl border border-dashed border-border-strong text-subtle"
                >
                  <Plus className="size-5" aria-hidden="true" />
                  <span className="text-xs">Empty slot</span>
                </li>
              );
            }
            const ToolIcon = toolIcons[loadoutSlot.tool.key];
            return (
              <li
                key={loadoutSlot.slot_number}
                className="min-h-28 rounded-xl border border-border bg-background/50 p-3"
              >
                <div className="flex items-center gap-3">
                  <span className="flex size-10 items-center justify-center rounded-lg bg-accent/10 text-accent">
                    <ToolIcon className="size-5" aria-hidden="true" />
                  </span>
                  <p className="text-sm font-semibold text-foreground">{loadoutSlot.tool.name}</p>
                </div>
                <div className="mt-4">
                  <div className="flex justify-between text-xs text-subtle">
                    <span>Mastery</span>
                    <span className="font-mono tabular-nums">{loadoutSlot.tool.mastery_percent}%</span>
                  </div>
                  <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-border">
                    <div
                      className="h-full rounded-full bg-linear-to-r from-amber-400 to-accent"
                      style={{ width: `${loadoutSlot.tool.mastery_percent}%` }}
                    />
                  </div>
                </div>
              </li>
            );
          })}
        </ul>
      )}
    </HubPanel>
  );
}
