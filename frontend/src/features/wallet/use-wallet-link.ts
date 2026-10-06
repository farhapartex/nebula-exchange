"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useSignMessage } from "wagmi";

import { createWalletChallenge, linkWallet, linkedWalletsQueryKey } from "@/features/wallet/api/wallet-api";
import { buildWalletLinkMessage } from "@/features/wallet/build-wallet-link-message";
import { gameChain } from "@/lib/web3/chain-config";

export function useWalletLink() {
  const queryClient = useQueryClient();
  const signMessage = useSignMessage();

  return useMutation({
    mutationFn: async (address: `0x${string}`) => {
      const walletChallenge = await createWalletChallenge(address, gameChain.id);
      const message = buildWalletLinkMessage({
        address,
        chainID: gameChain.id,
        nonce: walletChallenge.nonce,
        pageOrigin: window.location.origin,
        issuedAt: new Date(),
      });
      const signature = await signMessage.mutateAsync({ account: address, message });
      return linkWallet({ message, signature });
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: linkedWalletsQueryKey }),
  });
}
