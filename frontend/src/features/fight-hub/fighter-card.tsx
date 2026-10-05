"use client";

import Image from "next/image";

import { Skeleton } from "@/components/ui/skeleton";
import { useCurrentPlayer } from "@/features/auth/session/use-current-player";
import { rankTitleForLevel } from "@/features/fight-hub/fighter-ranks";
import { HubPanel } from "@/features/fight-hub/hub-panel";
import { useFighterProfile } from "@/features/fight-hub/use-fight-hub";

export function FighterCard() {
  const currentPlayer = useCurrentPlayer().data;
  const fighter = useFighterProfile().data;

  if (!currentPlayer || !fighter) {
    return <Skeleton className="h-full min-h-80 rounded-2xl" />;
  }

  const experienceProgress = Math.min(fighter.experience / fighter.experience_to_next_level, 1);

  return (
    <HubPanel className="overflow-hidden p-0">
      <div className="relative h-56">
        <Image
          src="/landing/fighter.webp"
          alt="Your fighter"
          fill
          sizes="22rem"
          className="object-cover object-[50%_12%]"
        />
        <div className="absolute inset-0 bg-linear-to-t from-surface via-surface/20 to-transparent" />
        <span className="absolute top-3 left-3 rounded-md border border-accent/40 bg-background/70 px-2 py-0.5 font-display text-sm tracking-[0.08em] text-accent-soft">
          {rankTitleForLevel(currentPlayer.current_level)}
        </span>
      </div>
      <div className="space-y-4 p-5 pt-1">
        <div>
          <p className="font-display text-2xl tracking-[0.04em] text-foreground">{currentPlayer.name}</p>
          <p className="text-xs text-muted">Fighter level {currentPlayer.current_level}</p>
        </div>
        <div>
          <div className="flex justify-between text-xs text-subtle">
            <span>Experience</span>
            <span className="font-mono tabular-nums">
              {fighter.experience} / {fighter.experience_to_next_level}
            </span>
          </div>
          <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-border">
            <div
              className="h-full rounded-full bg-linear-to-r from-amber-400 to-accent"
              style={{ width: `${experienceProgress * 100}%` }}
            />
          </div>
        </div>
        <dl className="grid grid-cols-2 gap-3 text-center">
          <div className="rounded-xl border border-border bg-background/50 p-2.5">
            <dt className="text-xs text-subtle">Wins</dt>
            <dd className="font-display text-2xl text-up">{currentPlayer.total_win}</dd>
          </div>
          <div className="rounded-xl border border-border bg-background/50 p-2.5">
            <dt className="text-xs text-subtle">Losses</dt>
            <dd className="font-display text-2xl text-down">{currentPlayer.total_lose}</dd>
          </div>
        </dl>
      </div>
    </HubPanel>
  );
}
