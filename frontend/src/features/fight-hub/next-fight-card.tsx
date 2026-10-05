"use client";

import Link from "next/link";
import { Play } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { ChapterPath } from "@/features/fight-hub/chapter-path";
import { HubPanel } from "@/features/fight-hub/hub-panel";
import { activeChapterOf, currentLevelOf, levelRoute } from "@/features/fight-hub/story-progress";
import { useStoryProgress } from "@/features/fight-hub/use-fight-hub";

export function NextFightCard() {
  const storyQuery = useStoryProgress();

  if (!storyQuery.data) {
    return <Skeleton className="h-full min-h-80 rounded-2xl" />;
  }

  const activeChapter = activeChapterOf(storyQuery.data);
  const nextLevel = currentLevelOf(storyQuery.data);
  if (!activeChapter) {
    return null;
  }

  return (
    <HubPanel className="relative flex flex-col justify-between gap-8 overflow-hidden">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute -right-24 -bottom-24 size-80 rounded-full bg-accent/20 blur-3xl"
      />
      <div className="relative">
        <p className="text-xs font-semibold tracking-[0.2em] text-highlight uppercase">
          Chapter {activeChapter.number}
        </p>
        <h1 className="mt-1 font-display text-4xl tracking-[0.03em] text-foreground sm:text-5xl">
          {activeChapter.title}
        </h1>
      </div>

      <div className="relative">
        <ChapterPath levels={activeChapter.levels} />
      </div>

      {nextLevel && (
        <div className="relative flex flex-wrap items-end justify-between gap-6 rounded-xl border border-accent/30 bg-background/60 p-5">
          <div>
            <p className="text-xs tracking-[0.18em] text-accent-soft uppercase">
              Next fight · Level {nextLevel.number}
            </p>
            <p className="mt-1 font-display text-3xl tracking-[0.03em] text-foreground">{nextLevel.title}</p>
            <p className="mt-1 text-sm text-muted">{nextLevel.teaser}</p>
          </div>
          <Button asChild size="lg" className="px-8 font-display text-xl tracking-[0.1em]">
            <Link href={levelRoute(nextLevel.id)}>
              <Play className="size-5 fill-current" aria-hidden="true" />
              Fight
            </Link>
          </Button>
        </div>
      )}
    </HubPanel>
  );
}
