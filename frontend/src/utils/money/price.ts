import { microUnitDecimals, toMicroUnits, type MicroUnitInput } from "@/utils/money/micro-units";

export function fractionDigitsForTickSize(tickSize: MicroUnitInput): number {
  let remainingTickSize = toMicroUnits(tickSize);
  if (remainingTickSize <= 0n) {
    throw new RangeError("tick size must be positive");
  }

  let trailingZeroCount = 0;
  while (remainingTickSize % 10n === 0n && trailingZeroCount < microUnitDecimals) {
    remainingTickSize /= 10n;
    trailingZeroCount += 1;
  }
  return microUnitDecimals - trailingZeroCount;
}

export type PriceDirection = "up" | "down" | "flat";

export type PriceChange = {
  direction: PriceDirection;
  changeInBasisPoints: bigint;
};

export function calculatePriceChange(currentPrice: MicroUnitInput, referencePrice: MicroUnitInput): PriceChange {
  const current = toMicroUnits(currentPrice);
  const reference = toMicroUnits(referencePrice);
  if (reference <= 0n) {
    return { direction: "flat", changeInBasisPoints: 0n };
  }

  const changeInBasisPoints = ((current - reference) * 10_000n) / reference;
  const direction: PriceDirection = changeInBasisPoints > 0n ? "up" : changeInBasisPoints < 0n ? "down" : "flat";
  return { direction, changeInBasisPoints };
}

export function formatBasisPointsAsPercent(changeInBasisPoints: bigint): string {
  const isNegative = changeInBasisPoints < 0n;
  const absoluteBasisPoints = isNegative ? -changeInBasisPoints : changeInBasisPoints;
  const wholePercent = absoluteBasisPoints / 100n;
  const fractionPercent = (absoluteBasisPoints % 100n).toString().padStart(2, "0");
  const sign = changeInBasisPoints > 0n ? "+" : isNegative ? "−" : "";
  return `${sign}${wholePercent}.${fractionPercent}%`;
}
