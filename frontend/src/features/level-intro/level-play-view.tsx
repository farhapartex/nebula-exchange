"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { Spinner } from "@/components/ui/spinner";
import { fetchFightSetup, fightSetupQueryKey } from "@/features/fight/api/fight-setup-api";
import { FightArena } from "@/features/fight/fight-arena";
import { fetchLevelStory, levelStoryQueryKey } from "@/features/level-intro/api/level-story-api";
import { FightArenaPlaceholder } from "@/features/level-intro/fight-arena-placeholder";
import { arrangeLevelStory } from "@/features/level-intro/level-story";
import { StorySlideshow } from "@/features/level-intro/story-slideshow";
import { isApiError } from "@/lib/api/api-error";

type PlayPhase = "story" | "fighting";

function FullScreenSpinner() {
  return (
    <div className="flex min-h-svh items-center justify-center">
      <Spinner />
    </div>
  );
}

export function LevelPlayView({ levelID }: { levelID: string }) {
  const [playPhase, setPlayPhase] = useState<PlayPhase>("story");
  const storyQuery = useQuery({
    queryKey: levelStoryQueryKey(levelID),
    queryFn: () => fetchLevelStory(levelID),
    retry: false,
    staleTime: 0,
  });
  const fightQuery = useQuery({
    queryKey: fightSetupQueryKey(levelID),
    queryFn: () => fetchFightSetup(levelID),
    retry: false,
  });
  const levelStory = storyQuery.data ? arrangeLevelStory(storyQuery.data) : null;
  const hasNoStory =
    (isApiError(storyQuery.error) && storyQuery.error.statusCode === 404) || (storyQuery.isSuccess && !levelStory);

  if (playPhase === "fighting" || hasNoStory) {
    if (fightQuery.data) {
      return <FightArena setup={fightQuery.data} />;
    }
    if (fightQuery.isError) {
      return <FightArenaPlaceholder />;
    }
    return <FullScreenSpinner />;
  }
  if (!levelStory) {
    return <FullScreenSpinner />;
  }
  return <StorySlideshow levelStory={levelStory} onPlay={() => setPlayPhase("fighting")} />;
}
