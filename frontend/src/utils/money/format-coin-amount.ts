export function formatCoinAmount(wholeCoins: string | number | bigint): string {
  return BigInt(wholeCoins).toLocaleString("en-US");
}
