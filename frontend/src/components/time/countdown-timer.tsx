"use client";

import { cn } from "@/utils/class-names";
import { formatClockCountdown } from "@/utils/time/format-duration";
import { useNow } from "@/utils/time/use-now";

type CountdownTimerProps = {
  endsAt: string;
  finishedLabel?: string;
  className?: string;
};

export function CountdownTimer({ endsAt, finishedLabel = "Done", className }: CountdownTimerProps) {
  const now = useNow();
  const remainingSeconds = (new Date(endsAt).getTime() - now) / 1000;
  return (
    <span role="timer" className={cn("font-mono tabular-nums", className)}>
      {remainingSeconds > 0 ? formatClockCountdown(remainingSeconds) : finishedLabel}
    </span>
  );
}
