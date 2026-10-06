const relativeTimeFormat = new Intl.RelativeTimeFormat("en-US", { numeric: "auto" });

const timeUnits: { unit: Intl.RelativeTimeFormatUnit; milliseconds: number }[] = [
  { unit: "day", milliseconds: 24 * 60 * 60 * 1000 },
  { unit: "hour", milliseconds: 60 * 60 * 1000 },
  { unit: "minute", milliseconds: 60 * 1000 },
];

export function formatRelativeTime(isoTimestamp: string, now: Date = new Date()): string {
  const differenceInMilliseconds = new Date(isoTimestamp).getTime() - now.getTime();
  for (const timeUnit of timeUnits) {
    if (Math.abs(differenceInMilliseconds) >= timeUnit.milliseconds) {
      return relativeTimeFormat.format(Math.round(differenceInMilliseconds / timeUnit.milliseconds), timeUnit.unit);
    }
  }
  return differenceInMilliseconds >= 0 ? "in less than a minute" : "just now";
}
