"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { Spinner } from "@/components/ui/spinner";
import { currentPlayerQueryKey } from "@/features/auth/session/use-current-player";
import { fetchFightSetup, fightSetupQueryKey } from "@/features/fight/api/fight-setup-api";
import { startFightSession } from "@/features/fight/api/fight-session-api";
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
  const [fightSessionID, setFightSessionID] = useState<string | null>(null);
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
  const queryClient = useQueryClient();
  const startFightMutation = useMutation({
    mutationFn: () => startFightSession(levelID),
    onSuccess: (startedFight) => {
      void queryClient.invalidateQueries({ queryKey: currentPlayerQueryKey });
      setFightSessionID(startedFight.id);
      setPlayPhase("fighting");
    },
  });

  async function startNewFight(): Promise<boolean> {
    try {
      await startFightMutation.mutateAsync();
      return true;
    } catch {
      return false;
    }
  }
  const levelStory = storyQuery.data ? arrangeLevelStory(storyQuery.data) : null;
  const hasNoStory =
    (isApiError(storyQuery.error) && storyQuery.error.statusCode === 404) || (storyQuery.isSuccess && !levelStory);

  if (playPhase === "fighting" || hasNoStory) {
    if (fightQuery.data) {
      return (
        <FightArena
          key={fightQuery.data.level_id}
          setup={fightQuery.data}
          fightSessionID={fightSessionID}
          onRestartFight={fightSessionID ? startNewFight : undefined}
        />
      );
    }
    if (fightQuery.isError) {
      return <FightArenaPlaceholder />;
    }
    return <FullScreenSpinner />;
  }
  if (!levelStory) {
    return <FullScreenSpinner />;
  }
  return (
    <StorySlideshow
      levelStory={levelStory}
      onPlay={() => startFightMutation.mutate()}
      isStartingFight={startFightMutation.isPending}
      startFightError={startFightErrorMessage(startFightMutation.error)}
    />
  );
}

function startFightErrorMessage(error: unknown): string | null {
  if (!error) {
    return null;
  }
  return isApiError(error) ? error.message : "Could not start the fight. Try again.";
}
