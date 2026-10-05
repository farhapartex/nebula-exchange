const hexColorPattern = /^#(?:[0-9a-f]{3}|[0-9a-f]{4}|[0-9a-f]{6}|[0-9a-f]{8})$/i;

export function safeHexColor(candidateColor: string | null | undefined, fallbackColor: string): string {
  return candidateColor && hexColorPattern.test(candidateColor) ? candidateColor : fallbackColor;
}
