export const topupAmountsInNc = ["5", "10", "25", "50", "100"] as const;

export type TopupAmount = (typeof topupAmountsInNc)[number];

export function toMicroUnits(amountInNc: TopupAmount): string {
  return `${amountInNc}000000`;
}
