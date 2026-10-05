import type { StoryChapter, StoryLevel } from "@/features/fight-hub/api/fight-hub-types";

export function currentLevelOf(chapters: StoryChapter[]): StoryLevel | null {
  for (const chapter of chapters) {
    const currentLevel = chapter.levels.find((storyLevel) => storyLevel.status === "current");
    if (currentLevel) {
      return currentLevel;
    }
  }
  return null;
}

export function activeChapterOf(chapters: StoryChapter[]): StoryChapter | null {
  return (
    chapters.find((chapter) => chapter.levels.some((storyLevel) => storyLevel.status === "current")) ??
    chapters[0] ??
    null
  );
}

export function levelRoute(levelID: string): string {
  return `/fight/${levelID}`;
}
