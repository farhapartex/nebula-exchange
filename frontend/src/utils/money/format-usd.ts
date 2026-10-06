const centsPerDollar = 100n;

export function formatUsd(cents: bigint): string {
  const sign = cents < 0n ? "-" : "";
  const absoluteCents = cents < 0n ? -cents : cents;
  const dollars = (absoluteCents / centsPerDollar).toLocaleString("en-US");
  const remainingCents = (absoluteCents % centsPerDollar).toString().padStart(2, "0");
  return `${sign}USD ${dollars}.${remainingCents}`;
}
