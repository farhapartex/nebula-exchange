import { ContentSection } from "@/components/layout/content-section";

export function TypographySection() {
  return (
    <ContentSection
      title="Typography"
      description="Geist for interface text, Geist Mono with tabular figures for numbers."
    >
      <div className="space-y-4">
        <p className="text-3xl font-semibold tracking-tight text-foreground sm:text-4xl">Mine. Craft. Trade.</p>
        <p className="text-xl font-semibold text-foreground">Section heading</p>
        <p className="max-w-prose text-sm leading-relaxed text-foreground">
          Body text. Send a Scout with a Basic Drill to the Asteroid Belt and bring back Iron and Copper Ore.
        </p>
        <p className="text-sm text-muted">Secondary text for labels and descriptions.</p>
        <p className="text-xs text-subtle">Hint text for helper copy and timestamps.</p>
        <div className="flex flex-wrap gap-x-8 gap-y-2 font-mono text-sm tabular-nums">
          <span className="text-foreground">1,284.500000 NC</span>
          <span className="text-muted">0.0105 × 40</span>
          <span className="text-highlight">00:14:59</span>
        </div>
      </div>
    </ContentSection>
  );
}
