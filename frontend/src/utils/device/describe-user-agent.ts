const browserPatterns: [RegExp, string][] = [
  [/Edg\//, "Edge"],
  [/OPR\/|Opera/, "Opera"],
  [/Firefox\//, "Firefox"],
  [/Chrome\//, "Chrome"],
  [/Safari\//, "Safari"],
];

const operatingSystemPatterns: [RegExp, string][] = [
  [/iPhone|iPad|iPod/, "iOS"],
  [/Android/, "Android"],
  [/Mac OS X|Macintosh/, "macOS"],
  [/Windows/, "Windows"],
  [/Linux/, "Linux"],
];

function firstMatchingLabel(userAgent: string, patterns: [RegExp, string][]): string | null {
  const matchingPattern = patterns.find(([pattern]) => pattern.test(userAgent));
  return matchingPattern ? matchingPattern[1] : null;
}

export function describeUserAgent(userAgent: string): string {
  if (!userAgent.trim()) {
    return "Unknown device";
  }
  const browserLabel = firstMatchingLabel(userAgent, browserPatterns);
  const operatingSystemLabel = firstMatchingLabel(userAgent, operatingSystemPatterns);
  if (browserLabel && operatingSystemLabel) {
    return `${browserLabel} on ${operatingSystemLabel}`;
  }
  return browserLabel ?? operatingSystemLabel ?? "Unknown device";
}
