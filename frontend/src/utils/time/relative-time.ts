const relativeTimeFormatter = new Intl.RelativeTimeFormat("en", { numeric: "auto", style: "short" });

const timeUnitsInSeconds: { unit: Intl.RelativeTimeFormatUnit; seconds: number }[] = [
  { unit: "year", seconds: 31_536_000 },
  { unit: "month", seconds: 2_592_000 },
  { unit: "week", seconds: 604_800 },
  { unit: "day", seconds: 86_400 },
  { unit: "hour", seconds: 3_600 },
  { unit: "minute", seconds: 60 },
];

export function formatRelativeTime(timestamp: string | Date, now: Date = new Date()): string {
  const elapsedSeconds = Math.round((new Date(timestamp).getTime() - now.getTime()) / 1000);
  if (Math.abs(elapsedSeconds) < 45) {
    return "just now";
  }

  for (const { unit, seconds } of timeUnitsInSeconds) {
    if (Math.abs(elapsedSeconds) >= seconds) {
      return relativeTimeFormatter.format(Math.round(elapsedSeconds / seconds), unit);
    }
  }
  return relativeTimeFormatter.format(Math.round(elapsedSeconds / 60), "minute");
}
