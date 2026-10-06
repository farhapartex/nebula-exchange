"use client";

import Link from "next/link";
import { Lock, Play, Sparkles } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useCurrentPlayer } from "@/features/auth/session/use-current-player";
import type { NextLevel } from "@/features/fight-hub/api/next-level-api";
import { ChapterPath } from "@/features/fight-hub/chapter-path";
import { HubPanel } from "@/features/fight-hub/hub-panel";
import { levelRoute, markLevelsByProgress } from "@/features/fight-hub/story-progress";
import { useNextLevel, useStoryProgress } from "@/features/fight-hub/use-fight-hub";

export function NextFightCard() {
  const nextLevel = useNextLevel().data;
  const currentPlayer = useCurrentPlayer().data;
  const storyChapters = useStoryProgress().data;

  if (!nextLevel || !currentPlayer) {
    return <Skeleton className="h-full min-h-80 rounded-2xl" />;
  }

  const chapterLevels = storyChapters?.find((chapter) => chapter.number === nextLevel.chapter.number)?.levels;

  return (
    <HubPanel className="relative flex flex-col justify-between gap-8 overflow-hidden">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute -right-24 -bottom-24 size-80 rounded-full bg-accent/20 blur-3xl"
      />
      <div className="relative">
        <p className="text-xs font-semibold tracking-[0.2em] text-highlight uppercase">
          Chapter {nextLevel.chapter.number}
        </p>
        <h1 className="mt-1 font-display text-4xl tracking-[0.03em] text-foreground sm:text-5xl">
          {nextLevel.chapter.title ?? "Coming soon"}
        </h1>
      </div>

      {chapterLevels && (
        <div className="relative">
          <ChapterPath levels={markLevelsByProgress(chapterLevels, currentPlayer.story_level)} />
        </div>
      )}

      <NextFightBox nextLevel={nextLevel} />
    </HubPanel>
  );
}

function NextFightBox({ nextLevel }: { nextLevel: NextLevel }) {
  if (nextLevel.status === "COMING_SOON" || !nextLevel.level) {
    return (
      <div className="relative flex items-center gap-4 rounded-xl border border-border bg-background/60 p-5">
        <Sparkles className="size-6 shrink-0 text-highlight" aria-hidden="true" />
        <div>
          <p className="font-display text-2xl tracking-[0.03em] text-foreground">
            Chapter {nextLevel.chapter.number} is coming soon
          </p>
          <p className="mt-1 text-sm text-muted">
            You have won every fight there is for now. New fights are on the way.
          </p>
        </div>
      </div>
    );
  }

  const isLocked = nextLevel.status === "LOCKED";
  return (
    <div className="relative flex flex-wrap items-end justify-between gap-6 rounded-xl border border-accent/30 bg-background/60 p-5">
      <div>
        <p className="text-xs tracking-[0.18em] text-accent-soft uppercase">
          Next fight · Level {nextLevel.level.number}
        </p>
        <p className="mt-1 font-display text-3xl tracking-[0.03em] text-foreground">{nextLevel.level.title}</p>
        <p className="mt-1 text-sm text-muted">
          {isLocked ? `Unlock chapter ${nextLevel.chapter.number} to keep fighting.` : nextLevel.level.teaser}
        </p>
      </div>
      {isLocked ? (
        <Button size="lg" variant="secondary" disabled className="px-8 font-display text-xl tracking-[0.1em]">
          <Lock className="size-5" aria-hidden="true" />
          Locked
        </Button>
      ) : (
        <Button asChild size="lg" className="px-8 font-display text-xl tracking-[0.1em]">
          <Link href={levelRoute(nextLevel.level.id)}>
            <Play className="size-5 fill-current" aria-hidden="true" />
            Fight
          </Link>
        </Button>
      )}
    </div>
  );
}
