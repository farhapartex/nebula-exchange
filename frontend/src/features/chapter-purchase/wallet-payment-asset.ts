export type WalletPaymentAsset = "ETH" | "USDC";

export const ethDecimals = 18;
export const usdcDecimals = 6;
export const shownEthDecimals = 6;
export const shownUsdcDecimals = 2;
export const ethPriceMovementBufferBasisPoints = 100n;
export const basisPointsPerWhole = 10_000n;

export function ethToSendFor(quotedWei: bigint): bigint {
  return quotedWei + (quotedWei * ethPriceMovementBufferBasisPoints) / basisPointsPerWhole;
}
