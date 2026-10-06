"use client";

import { CircleCheck } from "lucide-react";
import { useSwitchChain } from "wagmi";

import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { ConnectedWalletChip } from "@/features/wallet/connected-wallet-chip";
import { useWalletLink } from "@/features/wallet/use-wallet-link";
import type { WalletReadiness } from "@/features/wallet/use-wallet-readiness";
import { WalletConnectorList } from "@/features/wallet/wallet-connector-list";
import { gameChain } from "@/lib/web3/chain-config";
import { publishToastEvent } from "@/lib/notifications/toast-events";
import { shortenAddress } from "@/utils/web3/shorten-address";

export function WalletReadinessPanel({ readiness }: { readiness: WalletReadiness }) {
  const switchChain = useSwitchChain();
  const walletLink = useWalletLink();

  function linkConnectedWallet(address: `0x${string}`) {
    walletLink.mutate(address, {
      onError: () =>
        publishToastEvent({
          tone: "info",
          title: "Wallet not linked",
          description: "The message was not signed. You can try again any time.",
        }),
    });
  }

  return (
    <div className="space-y-3 rounded-xl border border-border bg-background/60 p-4">
      {readiness.stage === "CHECKING" && (
        <p className="flex items-center gap-2 text-sm text-muted">
          <Spinner /> Checking your wallet
        </p>
      )}

      {readiness.stage === "DISCONNECTED" && (
        <>
          <p className="text-sm text-muted">Connect a wallet to pay with USDC.</p>
          <WalletConnectorList />
        </>
      )}

      {readiness.stage === "WRONG_NETWORK" && (
        <>
          <ConnectedWalletChip address={readiness.address} statusLabel="Wrong network" />
          <p className="text-sm text-muted">Your wallet is on another network. Switch it to {gameChain.name} to pay.</p>
          <Button
            type="button"
            size="sm"
            isLoading={switchChain.isPending}
            onClick={() => switchChain.mutate({ chainId: gameChain.id })}
          >
            Switch to {gameChain.name}
          </Button>
        </>
      )}

      {readiness.stage === "NOT_LINKED" && (
        <>
          <ConnectedWalletChip address={readiness.address} statusLabel="Not linked" />
          <p className="text-sm text-muted">
            Link this wallet to your account first. You sign a message once. It is free and sends no transaction.
          </p>
          <Button
            type="button"
            size="sm"
            isLoading={walletLink.isPending}
            onClick={() => linkConnectedWallet(readiness.address)}
          >
            Link wallet
          </Button>
        </>
      )}

      {readiness.stage === "LINKED_TO_ANOTHER_WALLET" && (
        <>
          <ConnectedWalletChip address={readiness.address} statusLabel="Different wallet" />
          <p className="text-sm text-muted">
            Your account is linked to {shortenAddress(readiness.linkedAddress)}. Switch to that wallet in your wallet
            app to pay.
          </p>
        </>
      )}

      {readiness.stage === "READY" && (
        <>
          <ConnectedWalletChip address={readiness.address} statusLabel="Linked" />
          <p className="flex items-center gap-2 text-sm text-up">
            <CircleCheck className="size-4" aria-hidden="true" />
            Ready to pay with USDC on {gameChain.name}
          </p>
        </>
      )}
    </div>
  );
}
