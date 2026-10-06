import { formatUnits } from "viem";

export function formatTokenAmount(units: bigint, decimals: number, maximumFractionDigits: number): string {
  const [wholePart, fractionPart = ""] = formatUnits(units, decimals).split(".");
  const shownFraction = fractionPart.slice(0, maximumFractionDigits).replace(/0+$/, "");
  const roundedUp =
    fractionPart.length > maximumFractionDigits && /[1-9]/.test(fractionPart.slice(maximumFractionDigits));
  if (!roundedUp) {
    return shownFraction ? `${wholePart}.${shownFraction}` : wholePart;
  }
  const scale = 10n ** BigInt(maximumFractionDigits);
  const scaledUnits = (units * scale + 10n ** BigInt(decimals) - 1n) / 10n ** BigInt(decimals);
  return formatTokenAmount(scaledUnits, maximumFractionDigits, maximumFractionDigits);
}
