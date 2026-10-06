const centsPerDollar = 100n;

export function formatUsdcFromCents(cents: bigint): string {
  const dollars = (cents / centsPerDollar).toLocaleString("en-US");
  const remainingCents = (cents % centsPerDollar).toString().padStart(2, "0");
  return `${dollars}.${remainingCents} USDC`;
}
