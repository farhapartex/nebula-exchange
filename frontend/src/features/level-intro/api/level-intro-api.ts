import { requestData } from "@/lib/api/api-client";

export type SlidePalette = {
  background: string;
  glow: string;
  accent: string;
};

export type StorySlide = {
  id: string;
  eyebrow: string | null;
  heading: string;
  body: string;
  image: string | null;
  palette: SlidePalette;
};

export type LevelIntro = {
  level_id: string;
  chapter_number: number;
  level_number: number;
  level_title: string;
  slides: StorySlide[];
  call_to_action: {
    heading: string;
    body: string;
    button_label: string;
    image: string | null;
    palette: SlidePalette;
  };
};

export const levelIntroQueryKey = (levelID: string) => ["levels", levelID, "intro"] as const;

export function fetchLevelIntro(levelID: string): Promise<LevelIntro> {
  return requestData<LevelIntro>(`/levels/${encodeURIComponent(levelID)}/intro`);
}
