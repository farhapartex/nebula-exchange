import { requestAllPages } from "@/lib/api/api-client";

export type SlidePalette = {
  background: string;
  glow: string;
  accent: string;
};

export type StorySlideKind = "SLIDE" | "CALL_TO_ACTION";

export type StorySlide = {
  id: string;
  position: number;
  kind: StorySlideKind;
  eyebrow: string | null;
  heading: string;
  body: string;
  image: string | null;
  palette: SlidePalette;
  button_label: string | null;
};

export const levelStoryQueryKey = (levelID: string) => ["stories", levelID] as const;

export function fetchLevelStory(levelID: string): Promise<StorySlide[]> {
  return requestAllPages<StorySlide>("/stories", { level: levelID });
}
