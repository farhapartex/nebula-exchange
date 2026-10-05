"use client";

import { Pause } from "lucide-react";

import type { FighterSetup } from "@/features/fight/api/fight-setup-api";
import type { FighterHudState, FightSnapshot } from "@/features/fight/engine/fight-controller";
import { cn } from "@/utils/class-names";

function FighterBars({
  fighter,
  hud,
  alignment,
}: {
  fighter: FighterSetup;
  hud: FighterHudState;
  alignment: "start" | "end";
}) {
  const healthShare = Math.max(0, hud.health / hud.maxHealth);
  const staminaShare = Math.max(0, hud.stamina / hud.maxStamina);
  const isEnd = alignment === "end";
  return (
    <div className={cn("min-w-0 flex-1", isEnd && "text-right")}>
      <p className="truncate font-display text-xl leading-none tracking-[0.06em] text-foreground drop-shadow sm:text-2xl">
        {fighter.name}
      </p>
      <p className="truncate text-[0.6875rem] text-muted">{fighter.title}</p>
      <div
        className={cn(
          "mt-1.5 h-3.5 overflow-hidden rounded-sm border border-black/60 bg-black/50",
          isEnd && "flex justify-end",
        )}
      >
        <div
          className={cn(
            "h-full transition-[width] duration-150",
            healthShare > 0.3 ? "bg-linear-to-r from-amber-400 to-accent" : "bg-red-600",
          )}
          style={{ width: `${healthShare * 100}%` }}
        />
      </div>
      <div className={cn("mt-1 h-1.5 overflow-hidden rounded-sm bg-black/50", isEnd && "flex justify-end")}>
        <div className="h-full bg-highlight/80" style={{ width: `${staminaShare * 100}%` }} />
      </div>
    </div>
  );
}

type FightHudProps = {
  player: FighterSetup;
  enemy: FighterSetup;
  snapshot: FightSnapshot;
  onPause: () => void;
};

export function FightHud({ player, enemy, snapshot, onPause }: FightHudProps) {
  return (
    <div className="pointer-events-none absolute inset-x-0 top-0 z-10 flex items-start gap-3 p-3 sm:gap-6 sm:p-5">
      <FighterBars fighter={player} hud={snapshot.player} alignment="start" />
      <div className="flex shrink-0 flex-col items-center gap-1">
        <span
          className={cn(
            "font-display text-4xl leading-none tabular-nums drop-shadow sm:text-5xl",
            snapshot.remainingSeconds <= 10 ? "text-red-400" : "text-foreground",
          )}
        >
          {snapshot.remainingSeconds}
        </span>
        <button
          type="button"
          onClick={onPause}
          aria-label="Pause"
          className="pointer-events-auto flex size-8 items-center justify-center rounded-md bg-black/50 text-muted hover:text-foreground"
        >
          <Pause className="size-4" />
        </button>
      </div>
      <FighterBars fighter={enemy} hud={snapshot.enemy} alignment="end" />
    </div>
  );
}
