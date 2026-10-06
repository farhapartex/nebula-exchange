"use client";

import Link from "next/link";
import { RotateCcw, Trophy, Skull } from "lucide-react";

import { Button } from "@/components/ui/button";
import type { FightSnapshot } from "@/features/fight/engine/fight-controller";

const keyboardHints = [
  ["A / D", "Move"],
  ["J", "Punch"],
  ["K", "Kick"],
  ["L (hold)", "Block"],
  ["Space", "Dodge"],
  ["Esc", "Pause"],
];

export function ControlsHint() {
  return (
    <ul className="flex flex-wrap justify-center gap-x-4 gap-y-1 text-xs text-muted [@media(pointer:coarse)]:hidden">
      {keyboardHints.map(([key, action]) => (
        <li key={key}>
          <kbd className="rounded border border-border-strong bg-surface-raised px-1.5 py-0.5 font-mono text-[0.6875rem] text-foreground">
            {key}
          </kbd>{" "}
          {action}
        </li>
      ))}
    </ul>
  );
}

export function PauseOverlay({ onResume, onRestart }: { onResume: () => void; onRestart: () => void }) {
  return (
    <div className="absolute inset-0 z-20 flex flex-col items-center justify-center gap-6 bg-black/70 backdrop-blur-sm">
      <h2 className="font-display text-6xl tracking-[0.06em] text-foreground">Paused</h2>
      <div className="flex gap-3">
        <Button size="lg" onClick={onResume}>
          Resume
        </Button>
        <Button size="lg" variant="secondary" onClick={onRestart}>
          <RotateCcw className="size-4" />
          Restart
        </Button>
      </div>
      <ControlsHint />
      <Link href="/fight" className="text-sm text-muted hover:text-foreground">
        Leave the fight
      </Link>
    </div>
  );
}

type ResultOverlayProps = {
  snapshot: FightSnapshot;
  onRetry: () => void;
  isRetrying: boolean;
};

export function ResultOverlay({ snapshot, onRetry, isRetrying }: ResultOverlayProps) {
  const hasWon = snapshot.outcome === "won";
  const ResultIcon = hasWon ? Trophy : Skull;
  return (
    <div className="absolute inset-0 z-20 flex flex-col items-center justify-center gap-5 bg-black/70 px-6 text-center backdrop-blur-sm motion-safe:animate-pop-in">
      <ResultIcon className={hasWon ? "size-12 text-amber-400" : "size-12 text-red-400"} aria-hidden="true" />
      <h2 className="font-display text-6xl tracking-[0.04em] text-foreground sm:text-7xl">
        {hasWon ? "You survived" : "You fell"}
      </h2>
      <p className="max-w-md text-muted">
        {hasWon ? "The first bandit is down. The fire is still spreading." : "Get up. You are not done yet."}
      </p>
      <dl className="flex gap-8 text-sm">
        <div>
          <dt className="text-subtle">Damage dealt</dt>
          <dd className="font-display text-3xl text-foreground">{snapshot.damageDealt}</dd>
        </div>
        <div>
          <dt className="text-subtle">Damage taken</dt>
          <dd className="font-display text-3xl text-foreground">{snapshot.damageTaken}</dd>
        </div>
      </dl>
      <div className="flex gap-3">
        <Button size="lg" onClick={onRetry} isLoading={isRetrying}>
          {!isRetrying && <RotateCcw className="size-4" />}
          {hasWon ? "Fight again" : "Try again"}
        </Button>
        <Button asChild size="lg" variant="secondary">
          <Link href="/fight">Back to the hub</Link>
        </Button>
      </div>
    </div>
  );
}
