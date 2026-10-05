import { Check, Lock, Trophy } from "lucide-react";

import type { StoryLevel } from "@/features/fight-hub/api/fight-hub-types";
import { cn } from "@/utils/class-names";

export function ChapterPath({ levels }: { levels: StoryLevel[] }) {
  return (
    <ol className="flex items-center" aria-label="Chapter progress">
      {levels.map((storyLevel, levelIndex) => {
        const isFinale = storyLevel.number === null;
        const isLast = levelIndex === levels.length - 1;
        return (
          <li key={storyLevel.id} className={cn("flex items-center", !isLast && "flex-1")}>
            <div className="flex flex-col items-center gap-1.5">
              <span
                title={storyLevel.title}
                className={cn(
                  "flex size-9 items-center justify-center rounded-full border-2 font-display text-base transition-colors",
                  storyLevel.status === "completed" && "border-accent bg-accent text-background",
                  storyLevel.status === "current" &&
                    "border-accent bg-accent/15 text-accent-soft shadow-[0_0_18px_-2px] shadow-accent motion-safe:animate-pulse",
                  storyLevel.status === "locked" && "border-border-strong bg-background/60 text-subtle",
                )}
              >
                {storyLevel.status === "completed" ? (
                  <Check className="size-4" aria-label="Completed" />
                ) : isFinale ? (
                  <Trophy className="size-4" aria-hidden="true" />
                ) : storyLevel.status === "locked" ? (
                  <Lock className="size-3.5" aria-hidden="true" />
                ) : (
                  storyLevel.number
                )}
              </span>
              <span
                className={cn("text-[0.6875rem]", storyLevel.status === "current" ? "text-accent-soft" : "text-subtle")}
              >
                {isFinale ? "Finale" : `Level ${storyLevel.number}`}
              </span>
            </div>
            {!isLast && (
              <span
                aria-hidden="true"
                className={cn(
                  "mx-1 mb-5 h-0.5 flex-1 rounded-full",
                  storyLevel.status === "completed" ? "bg-accent" : "bg-border-strong",
                )}
              />
            )}
          </li>
        );
      })}
    </ol>
  );
}
