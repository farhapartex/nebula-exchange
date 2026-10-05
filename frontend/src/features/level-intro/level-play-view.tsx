"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { Spinner } from "@/components/ui/spinner";
import { fetchFightSetup, fightSetupQueryKey } from "@/features/fight/api/fight-setup-api";
import { FightArena } from "@/features/fight/fight-arena";
import { fetchLevelIntro, levelIntroQueryKey } from "@/features/level-intro/api/level-intro-api";
import { FightArenaPlaceholder } from "@/features/level-intro/fight-arena-placeholder";
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
  const introQuery = useQuery({
    queryKey: levelIntroQueryKey(levelID),
    queryFn: () => fetchLevelIntro(levelID),
    retry: false,
  });
  const fightQuery = useQuery({
    queryKey: fightSetupQueryKey(levelID),
    queryFn: () => fetchFightSetup(levelID),
    retry: false,
  });
  const hasNoStory = isApiError(introQuery.error) && introQuery.error.statusCode === 404;

  if (playPhase === "fighting" || hasNoStory) {
    if (fightQuery.data) {
      return <FightArena setup={fightQuery.data} />;
    }
    if (fightQuery.isError) {
      return <FightArenaPlaceholder levelTitle={introQuery.data?.level_title ?? "Fight"} />;
    }
    return <FullScreenSpinner />;
  }
  if (!introQuery.data) {
    return <FullScreenSpinner />;
  }
  return <StorySlideshow levelIntro={introQuery.data} onPlay={() => setPlayPhase("fighting")} />;
}
