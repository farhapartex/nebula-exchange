const basisPointsPerWhole = 10_000;

export function formatLootChance(chanceBasisPoints: number): string {
  if (chanceBasisPoints >= basisPointsPerWhole) {
    return "Always";
  }
  const percent = (chanceBasisPoints / basisPointsPerWhole) * 100;
  return `${Number.isInteger(percent) ? percent : percent.toFixed(2)}% chance`;
}

export function formatQuantityRange(minimumQuantity: number, maximumQuantity: number): string {
  return minimumQuantity === maximumQuantity ? `${minimumQuantity}` : `${minimumQuantity}–${maximumQuantity}`;
}
