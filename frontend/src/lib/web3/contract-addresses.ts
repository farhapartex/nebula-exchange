import { getAddress, isAddress, type Address } from "viem";

function readContractAddress(configuredAddress: string | undefined): Address | null {
  return configuredAddress && isAddress(configuredAddress) ? getAddress(configuredAddress) : null;
}

export const chapterPaymentVaultAddress = readContractAddress(process.env.NEXT_PUBLIC_CHAPTER_PAYMENT_VAULT_ADDRESS);
export const paymentTokenAddress = readContractAddress(process.env.NEXT_PUBLIC_PAYMENT_TOKEN_ADDRESS);
