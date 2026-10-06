"use client";

import { useQuery } from "@tanstack/react-query";
import { useConnection } from "wagmi";

import { fetchLinkedWallets, linkedWalletsQueryKey } from "@/features/wallet/api/wallet-api";
import { gameChain } from "@/lib/web3/chain-config";

export type WalletReadiness =
  | { stage: "CHECKING" }
  | { stage: "DISCONNECTED" }
  | { stage: "WRONG_NETWORK"; address: `0x${string}` }
  | { stage: "NOT_LINKED"; address: `0x${string}` }
  | { stage: "LINKED_TO_ANOTHER_WALLET"; address: `0x${string}`; linkedAddress: string }
  | { stage: "READY"; address: `0x${string}` };

export function useWalletReadiness(isEnabled: boolean): WalletReadiness {
  const connection = useConnection();
  const linkedWalletsQuery = useQuery({
    queryKey: linkedWalletsQueryKey,
    queryFn: fetchLinkedWallets,
    enabled: isEnabled,
  });

  if (connection.status === "connecting" || connection.status === "reconnecting") {
    return { stage: "CHECKING" };
  }
  if (connection.status !== "connected" || !connection.address) {
    return { stage: "DISCONNECTED" };
  }
  const address = connection.address;
  if (connection.chainId !== gameChain.id) {
    return { stage: "WRONG_NETWORK", address };
  }
  if (!linkedWalletsQuery.data) {
    return { stage: "CHECKING" };
  }
  const linkedWallet = linkedWalletsQuery.data[0];
  if (!linkedWallet) {
    return { stage: "NOT_LINKED", address };
  }
  if (linkedWallet.address.toLowerCase() !== address.toLowerCase()) {
    return { stage: "LINKED_TO_ANOTHER_WALLET", address, linkedAddress: linkedWallet.address };
  }
  return { stage: "READY", address };
}
