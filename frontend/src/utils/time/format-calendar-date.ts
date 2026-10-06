const calendarDateFormat = new Intl.DateTimeFormat("en-US", { year: "numeric", month: "short", day: "numeric" });

export function formatCalendarDate(isoTimestamp: string): string {
  return calendarDateFormat.format(new Date(isoTimestamp));
}
