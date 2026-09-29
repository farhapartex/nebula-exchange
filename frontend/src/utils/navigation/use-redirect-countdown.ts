"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

export function useRedirectCountdown(destinationPath: string, totalSeconds: number, isEnabled: boolean): number {
  const router = useRouter();
  const [remainingSeconds, setRemainingSeconds] = useState(totalSeconds);

  useEffect(() => {
    if (!isEnabled) {
      return;
    }
    const countdownInterval = window.setInterval(() => {
      setRemainingSeconds((previousSeconds) => Math.max(previousSeconds - 1, 0));
    }, 1000);
    const redirectTimeout = window.setTimeout(() => router.replace(destinationPath), totalSeconds * 1000);

    return () => {
      window.clearInterval(countdownInterval);
      window.clearTimeout(redirectTimeout);
    };
  }, [destinationPath, isEnabled, router, totalSeconds]);

  return remainingSeconds;
}
