"use client";

import { useEffect, useState } from "react";

export function useNow(tickIntervalInMilliseconds = 1_000): number {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const tickInterval = window.setInterval(() => setNow(Date.now()), tickIntervalInMilliseconds);
    return () => window.clearInterval(tickInterval);
  }, [tickIntervalInMilliseconds]);
  return now;
}
