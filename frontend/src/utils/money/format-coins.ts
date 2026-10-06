const microUnitsPerCoin = 1_000_000n;

export function formatCoins(microUnits: string): string {
  const amount = BigInt(microUnits);
  const sign = amount < 0n ? "-" : "";
  const absoluteAmount = amount < 0n ? -amount : amount;
  const wholeCoins = (absoluteAmount / microUnitsPerCoin).toLocaleString("en-US");
  const remainder = absoluteAmount % microUnitsPerCoin;
  if (remainder === 0n) {
    return `${sign}${wholeCoins}`;
  }
  const fraction = remainder.toString().padStart(6, "0").replace(/0+$/, "");
  return `${sign}${wholeCoins}.${fraction}`;
}
