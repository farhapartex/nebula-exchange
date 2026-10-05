"use client";

import { Lock } from "lucide-react";

import { HubPanel } from "@/features/fight-hub/hub-panel";
import { useStoryProgress } from "@/features/fight-hub/use-fight-hub";

export function LockedChapterCard() {
  const storyQuery = useStoryProgress();
  const nextLockedChapter = storyQuery.data?.find((chapter) => !chapter.is_unlocked);
  if (!nextLockedChapter) {
    return null;
  }
  return (
    <HubPanel className="relative overflow-hidden">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute -top-16 -right-16 size-48 rounded-full bg-highlight/10 blur-3xl"
      />
      <div className="relative flex items-start gap-4">
        <span className="flex size-11 shrink-0 items-center justify-center rounded-xl border border-border-strong bg-background/60 text-muted">
          <Lock className="size-5" aria-hidden="true" />
        </span>
        <div>
          <p className="text-xs tracking-[0.18em] text-subtle uppercase">Chapter {nextLockedChapter.number}</p>
          <p className="font-display text-2xl tracking-[0.04em] text-foreground">{nextLockedChapter.title}</p>
          <p className="mt-1 text-sm text-muted">Finish Chapter {nextLockedChapter.number - 1} to unlock.</p>
        </div>
      </div>
    </HubPanel>
  );
}
