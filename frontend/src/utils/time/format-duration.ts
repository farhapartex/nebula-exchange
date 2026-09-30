export function formatCountdown(totalSeconds: number): string {
  const clampedSeconds = Math.max(Math.ceil(totalSeconds), 0);
  const minutes = Math.floor(clampedSeconds / 60);
  const seconds = clampedSeconds % 60;
  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

export function describeWaitTime(totalSeconds: number): string {
  if (totalSeconds < 60) {
    const seconds = Math.max(Math.ceil(totalSeconds), 1);
    return seconds === 1 ? "1 second" : `${seconds} seconds`;
  }
  const minutes = Math.ceil(totalSeconds / 60);
  return minutes === 1 ? "1 minute" : `${minutes} minutes`;
}

export function formatDurationShort(totalSeconds: number): string {
  if (totalSeconds < 60) {
    return `${totalSeconds}s`;
  }
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  if (hours === 0) {
    return `${minutes} min`;
  }
  return minutes === 0 ? `${hours} h` : `${hours} h ${minutes} min`;
}
