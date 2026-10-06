import { requestAllPages, requestData } from "@/lib/api/api-client";

export type LinkedWallet = {
  address: string;
  chain_id: number;
  linked_at: string;
};

export type WalletChallenge = {
  nonce: string;
  expires_at: string;
};

export type WalletLinkRequest = {
  message: string;
  signature: string;
};

export const linkedWalletsQueryKey = ["wallets"] as const;

export function fetchLinkedWallets(): Promise<LinkedWallet[]> {
  return requestAllPages<LinkedWallet>("/wallets");
}

export function createWalletChallenge(address: string, chainID: number): Promise<WalletChallenge> {
  return requestData<WalletChallenge>("/wallet-challenges", {
    method: "POST",
    body: { address, chain_id: chainID },
  });
}

export function linkWallet(walletLinkRequest: WalletLinkRequest): Promise<LinkedWallet> {
  return requestData<LinkedWallet>("/wallets", { method: "POST", body: walletLinkRequest });
}
