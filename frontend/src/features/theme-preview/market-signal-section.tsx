import { ContentSection } from "@/components/layout/content-section";
import { marketSignalExamples, type MarketSignalExample } from "@/features/theme-preview/theme-tokens";
import { cn } from "@/utils/class-names";

const directionStyles: Record<MarketSignalExample["direction"], { arrow: string; className: string }> = {
  up: { arrow: "▲", className: "bg-up-soft/40 text-up" },
  down: { arrow: "▼", className: "bg-down-soft/40 text-down" },
  flat: { arrow: "■", className: "bg-surface-raised text-muted" },
};

export function MarketSignalSection() {
  return (
    <ContentSection
      title="Market signals"
      description="Up is green and down is red, and every color also carries an arrow for colorblind players."
    >
      <ul className="divide-y divide-border overflow-hidden rounded-xl border border-border">
        {marketSignalExamples.map((example) => {
          const directionStyle = directionStyles[example.direction];
          return (
            <li key={example.symbol} className="flex items-center justify-between gap-4 bg-background/60 px-4 py-3">
              <span className="font-mono text-sm font-medium text-foreground">{example.symbol}</span>
              <span className="flex items-center gap-4 font-mono text-sm tabular-nums">
                <span className="text-foreground">{example.lastPrice}</span>
                <span
                  className={cn(
                    "inline-flex min-w-24 items-center justify-end gap-1.5 rounded-md px-2 py-1",
                    directionStyle.className,
                  )}
                >
                  <span aria-hidden="true" className="text-[0.65rem]">
                    {directionStyle.arrow}
                  </span>
                  {example.changePercent}
                </span>
              </span>
            </li>
          );
        })}
      </ul>
    </ContentSection>
  );
}
