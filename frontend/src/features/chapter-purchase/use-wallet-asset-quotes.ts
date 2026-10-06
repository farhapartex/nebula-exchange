"use client";

import { erc20Abi } from "viem";
import { useBalance, useReadContract } from "wagmi";

import { ethToSendFor } from "@/features/chapter-purchase/wallet-payment-asset";
import { chapterPaymentVaultAbi } from "@/lib/web3/abi/chapter-payment-vault-abi";
import { chapterPaymentVaultAddress, paymentTokenAddress } from "@/lib/web3/contract-addresses";

const quoteRefreshInMilliseconds = 15_000;

export type WalletAssetQuote = {
  priceUnits: bigint | null;
  unitsToSend: bigint | null;
  balanceUnits: bigint | null;
  isUnavailable: boolean;
  hasEnoughBalance: boolean;
};

export function useWalletAssetQuotes(totalCents: bigint, walletAddress: `0x${string}` | undefined) {
  const isConfigured = chapterPaymentVaultAddress !== null && paymentTokenAddress !== null;
  const ethQuoteQuery = useReadContract({
    address: chapterPaymentVaultAddress ?? undefined,
    abi: chapterPaymentVaultAbi,
    functionName: "quoteWei",
    args: [totalCents],
    query: { enabled: isConfigured, refetchInterval: quoteRefreshInMilliseconds },
  });
  const usdcQuoteQuery = useReadContract({
    address: chapterPaymentVaultAddress ?? undefined,
    abi: chapterPaymentVaultAbi,
    functionName: "quoteTokenUnits",
    args: [totalCents],
    query: { enabled: isConfigured },
  });
  const ethBalanceQuery = useBalance({ address: walletAddress, query: { enabled: walletAddress !== undefined } });
  const usdcBalanceQuery = useReadContract({
    address: paymentTokenAddress ?? undefined,
    abi: erc20Abi,
    functionName: "balanceOf",
    args: walletAddress ? [walletAddress] : undefined,
    query: { enabled: isConfigured && walletAddress !== undefined },
  });

  const ethPrice = ethQuoteQuery.data ?? null;
  const ethToSend = ethPrice === null ? null : ethToSendFor(ethPrice);
  const ethBalance = ethBalanceQuery.data?.value ?? null;
  const usdcPrice = usdcQuoteQuery.data ?? null;
  const usdcBalance = usdcBalanceQuery.data ?? null;

  const ethQuote: WalletAssetQuote = {
    priceUnits: ethPrice,
    unitsToSend: ethToSend,
    balanceUnits: ethBalance,
    isUnavailable: !isConfigured || ethQuoteQuery.isError,
    hasEnoughBalance: ethToSend !== null && ethBalance !== null && ethBalance >= ethToSend,
  };
  const usdcQuote: WalletAssetQuote = {
    priceUnits: usdcPrice,
    unitsToSend: usdcPrice,
    balanceUnits: usdcBalance,
    isUnavailable: !isConfigured || usdcQuoteQuery.isError,
    hasEnoughBalance: usdcPrice !== null && usdcBalance !== null && usdcBalance >= usdcPrice,
  };
  return { ethQuote, usdcQuote, isLoading: ethQuoteQuery.isLoading || usdcQuoteQuery.isLoading };
}
