import type { StorySlide } from "@/features/level-intro/api/level-story-api";

export type LevelStory = {
  slides: StorySlide[];
  callToAction: StorySlide & { button_label: string };
};

export function arrangeLevelStory(storySlides: StorySlide[]): LevelStory | null {
  const orderedSlides = [...storySlides].sort((first, second) => first.position - second.position);
  const callToAction = orderedSlides.find((storySlide) => storySlide.kind === "CALL_TO_ACTION");
  const slides = orderedSlides.filter((storySlide) => storySlide.kind === "SLIDE");
  if (!callToAction?.button_label || slides.length === 0) {
    return null;
  }
  return { slides, callToAction: { ...callToAction, button_label: callToAction.button_label } };
}
