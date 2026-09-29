export const microUnitDecimals = 6;
export const microUnitsPerNc = 1_000_000n;

export type MicroUnitInput = bigint | string;

const integerPattern = /^-?\d+$/;
const decimalInputPattern = /^\d*(\.\d*)?$/;

export function toMicroUnits(amount: MicroUnitInput): bigint {
  if (typeof amount === "bigint") {
    return amount;
  }
  if (!integerPattern.test(amount)) {
    throw new Error(`"${amount}" is not a micro-unit integer string`);
  }
  return BigInt(amount);
}

type FormatMicroUnitsOptions = {
  fractionDigits?: number;
  useGrouping?: boolean;
  showPlusSign?: boolean;
};

export function formatMicroUnits(amount: MicroUnitInput, options: FormatMicroUnitsOptions = {}): string {
  const { fractionDigits = microUnitDecimals, useGrouping = true, showPlusSign = false } = options;
  if (fractionDigits < 0 || fractionDigits > microUnitDecimals) {
    throw new RangeError(`fractionDigits must be between 0 and ${microUnitDecimals}`);
  }

  const microUnits = toMicroUnits(amount);
  const isNegative = microUnits < 0n;
  const absoluteMicroUnits = isNegative ? -microUnits : microUnits;

  const truncationDivisor = 10n ** BigInt(microUnitDecimals - fractionDigits);
  const truncatedMicroUnits = (absoluteMicroUnits / truncationDivisor) * truncationDivisor;

  const wholePart = truncatedMicroUnits / microUnitsPerNc;
  const fractionPart = (truncatedMicroUnits % microUnitsPerNc)
    .toString()
    .padStart(microUnitDecimals, "0")
    .slice(0, fractionDigits);

  const wholeText = useGrouping ? groupThousands(wholePart.toString()) : wholePart.toString();
  const unsignedText = fractionDigits > 0 ? `${wholeText}.${fractionPart}` : wholeText;
  const isDisplayedAsZero = truncatedMicroUnits === 0n;

  if (isNegative && !isDisplayedAsZero) {
    return `−${unsignedText}`;
  }
  if (showPlusSign && !isDisplayedAsZero) {
    return `+${unsignedText}`;
  }
  return unsignedText;
}

export function parseNcToMicroUnits(decimalText: string): bigint | null {
  const normalizedText = decimalText.trim().replaceAll(",", "");
  if (normalizedText === "" || normalizedText === "." || !decimalInputPattern.test(normalizedText)) {
    return null;
  }

  const [wholeText = "", fractionText = ""] = normalizedText.split(".");
  if (fractionText.length > microUnitDecimals) {
    return null;
  }

  const wholeMicroUnits = BigInt(wholeText || "0") * microUnitsPerNc;
  const fractionMicroUnits = BigInt(fractionText.padEnd(microUnitDecimals, "0") || "0");
  return wholeMicroUnits + fractionMicroUnits;
}

function groupThousands(digits: string): string {
  return digits.replace(/\B(?=(\d{3})+(?!\d))/g, ",");
}
