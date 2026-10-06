type MasteryMeterProps = {
  masteryLevel: number;
  maxMasteryLevel: number;
};

export function MasteryMeter({ masteryLevel, maxMasteryLevel }: MasteryMeterProps) {
  const filledPercent = Math.round((masteryLevel / maxMasteryLevel) * 100);
  return (
    <div>
      <div className="flex justify-between text-xs text-subtle">
        <span>Mastery</span>
        <span className="font-mono text-foreground tabular-nums">
          {masteryLevel} / {maxMasteryLevel}
        </span>
      </div>
      <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-border">
        <div
          className="h-full rounded-full bg-linear-to-r from-amber-400 to-accent"
          style={{ width: `${filledPercent}%` }}
        />
      </div>
    </div>
  );
}
