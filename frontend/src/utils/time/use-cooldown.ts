"use client";

import { useCallback, useEffect, useState } from "react";

export function useCooldown(durationInSeconds: number) {
  const [cooldownEndsAt, setCooldownEndsAt] = useState<number | null>(null);
  const [remainingSeconds, setRemainingSeconds] = useState(0);

  useEffect(() => {
    if (cooldownEndsAt === null) {
      return;
    }
    const updateRemainingSeconds = () => {
      const secondsLeft = Math.max(Math.ceil((cooldownEndsAt - Date.now()) / 1000), 0);
      setRemainingSeconds(secondsLeft);
      if (secondsLeft === 0) {
        setCooldownEndsAt(null);
      }
    };
    updateRemainingSeconds();
    const tickInterval = window.setInterval(updateRemainingSeconds, 250);
    return () => window.clearInterval(tickInterval);
  }, [cooldownEndsAt]);

  const startCooldown = useCallback(
    (overrideDurationInSeconds?: number) => {
      const cooldownSeconds = overrideDurationInSeconds ?? durationInSeconds;
      setCooldownEndsAt(Date.now() + cooldownSeconds * 1000);
      setRemainingSeconds(cooldownSeconds);
    },
    [durationInSeconds],
  );

  return { remainingSeconds, isCoolingDown: remainingSeconds > 0, startCooldown };
}
