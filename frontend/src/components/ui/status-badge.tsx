import type { Tone } from "@/components/ui/tone";
import { cn } from "@/utils/class-names";

const toneClassNames: Record<Tone, { badge: string; dot: string }> = {
  neutral: { badge: "border-border-strong bg-surface-raised text-muted", dot: "bg-muted" },
  accent: { badge: "border-accent/40 bg-accent/15 text-accent-soft", dot: "bg-accent" },
  info: { badge: "border-info/40 bg-info/10 text-info", dot: "bg-info" },
  success: { badge: "border-up/40 bg-up/10 text-up", dot: "bg-up" },
  warning: { badge: "border-warning/40 bg-warning/10 text-warning", dot: "bg-warning" },
  danger: { badge: "border-down/40 bg-down/10 text-down", dot: "bg-down" },
};

type StatusBadgeProps = {
  label: string;
  tone?: Tone;
  isPulsing?: boolean;
  className?: string;
};

export function StatusBadge({ label, tone = "neutral", isPulsing = false, className }: StatusBadgeProps) {
  const toneClassName = toneClassNames[tone];

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-xs font-medium whitespace-nowrap",
        toneClassName.badge,
        className,
      )}
    >
      <span className={cn("size-1.5 rounded-full", toneClassName.dot, isPulsing && "animate-pulse")} />
      {label}
    </span>
  );
}
