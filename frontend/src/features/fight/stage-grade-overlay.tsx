import type { StageSetup } from "@/features/fight/api/fight-setup-api";
import { safeHexColor } from "@/utils/colors/safe-hex-color";

const grainTexture =
  "url(\"data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='160' height='160'><filter id='grain'><feTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='2' stitchTiles='stitch'/></filter><rect width='100%' height='100%' filter='url(%23grain)'/></svg>\")";

export function StageGradeOverlay({ grade }: { grade: StageSetup["grade"] }) {
  return (
    <div aria-hidden="true" className="pointer-events-none absolute inset-0">
      <div
        className="absolute inset-0 mix-blend-soft-light"
        style={{ backgroundColor: safeHexColor(grade.tint, "#ff6a00"), opacity: 0.35 }}
      />
      <div
        className="absolute inset-0"
        style={{
          background: `radial-gradient(ellipse at 50% 45%, transparent 45%, rgba(0, 0, 0, ${grade.vignette}) 100%)`,
        }}
      />
      <div
        className="absolute inset-0 mix-blend-overlay"
        style={{ backgroundImage: grainTexture, opacity: grade.grain * 4 }}
      />
    </div>
  );
}
