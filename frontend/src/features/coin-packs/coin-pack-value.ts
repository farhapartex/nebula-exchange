import type { CoinPack } from "@/features/coin-packs/api/coin-pack-api";

export function extraCoinsPercent(coinPack: CoinPack, baseCoinPack: CoinPack): number {
  const baseCoinsTimesPrice = BigInt(baseCoinPack.coins) * BigInt(coinPack.price_cents);
  if (baseCoinsTimesPrice === 0n) {
    return 0;
  }
  const extraPercentTimesBase =
    (BigInt(coinPack.coins) * BigInt(baseCoinPack.price_cents) - baseCoinsTimesPrice) * 100n;
  const roundedPercent = (extraPercentTimesBase + baseCoinsTimesPrice / 2n) / baseCoinsTimesPrice;
  return roundedPercent > 0n ? Number(roundedPercent) : 0;
}

export function cheapestCoinPack(coinPacks: CoinPack[]): CoinPack | undefined {
  return coinPacks.reduce<CoinPack | undefined>(
    (cheapest, coinPack) =>
      !cheapest || BigInt(coinPack.price_cents) < BigInt(cheapest.price_cents) ? coinPack : cheapest,
    undefined,
  );
}

export function smallestCoinPackCovering(coinPacks: CoinPack[], neededCoins: bigint): CoinPack | undefined {
  const coveringPacks = coinPacks.filter((coinPack) => BigInt(coinPack.coins) >= neededCoins);
  if (coveringPacks.length > 0) {
    return cheapestCoinPack(coveringPacks);
  }
  return coinPacks.reduce<CoinPack | undefined>(
    (biggest, coinPack) => (!biggest || BigInt(coinPack.coins) > BigInt(biggest.coins) ? coinPack : biggest),
    undefined,
  );
}
