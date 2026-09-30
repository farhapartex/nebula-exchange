export const basisPointsPerWhole = 10_000;

export function parseBasisPoints(decimalText: string | undefined): number | null {
  if (!decimalText) {
    return null;
  }
  const decimalMatch = /^(\d+)(?:\.(\d{1,4}))?$/.exec(decimalText.trim());
  if (!decimalMatch) {
    return null;
  }
  const fractionDigits = (decimalMatch[2] ?? "").padEnd(4, "0");
  return Number(decimalMatch[1]) * basisPointsPerWhole + Number(fractionDigits);
}

export function applyBasisPoints(value: number, basisPoints: number): number {
  return Math.floor((value * basisPoints) / basisPointsPerWhole);
}
