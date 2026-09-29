import { cn } from "@/utils/class-names";
import { formatBasisPointsAsPercent, type PriceDirection } from "@/utils/money/price";

const directionAppearance: Record<PriceDirection, { arrow: string; label: string; className: string }> = {
  up: { arrow: "▲", label: "up", className: "text-up" },
  down: { arrow: "▼", label: "down", className: "text-down" },
  flat: { arrow: "■", label: "unchanged", className: "text-muted" },
};

type PriceChangeProps = {
  changeInBasisPoints: bigint;
  variant?: "text" | "pill";
  className?: string;
};

function directionOf(changeInBasisPoints: bigint): PriceDirection {
  if (changeInBasisPoints > 0n) {
    return "up";
  }
  return changeInBasisPoints < 0n ? "down" : "flat";
}

export function PriceChange({ changeInBasisPoints, variant = "text", className }: PriceChangeProps) {
  const direction = directionOf(changeInBasisPoints);
  const appearance = directionAppearance[direction];
  const percentText = formatBasisPointsAsPercent(changeInBasisPoints);

  return (
    <span
      aria-label={`${appearance.label} ${percentText}`}
      className={cn(
        "inline-flex items-center gap-1 font-mono whitespace-nowrap tabular-nums",
        appearance.className,
        variant === "pill" && "rounded-md px-2 py-0.5",
        variant === "pill" && direction === "up" && "bg-up-soft/40",
        variant === "pill" && direction === "down" && "bg-down-soft/40",
        variant === "pill" && direction === "flat" && "bg-surface-raised",
        className,
      )}
    >
      <span aria-hidden="true" className="text-[0.65em]">
        {appearance.arrow}
      </span>
      {percentText}
    </span>
  );
}
