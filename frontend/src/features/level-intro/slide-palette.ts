import type { SlidePalette } from "@/features/level-intro/api/level-story-api";
import { safeHexColor } from "@/utils/colors/safe-hex-color";

const themePalette: SlidePalette = {
  background: "#0a0807",
  glow: "#f9731633",
  accent: "#f97316",
};

export function resolveSlidePalette(palette: Partial<SlidePalette> | undefined): SlidePalette {
  return {
    background: safeHexColor(palette?.background, themePalette.background),
    glow: safeHexColor(palette?.glow, themePalette.glow),
    accent: safeHexColor(palette?.accent, themePalette.accent),
  };
}
