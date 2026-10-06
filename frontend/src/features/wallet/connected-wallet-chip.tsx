"use client";

import { useDisconnect } from "wagmi";

import { shortenAddress } from "@/utils/web3/shorten-address";

export function ConnectedWalletChip({ address, statusLabel }: { address: string; statusLabel: string }) {
  const disconnect = useDisconnect();
  return (
    <div className="flex flex-wrap items-center justify-between gap-2">
      <span className="flex items-center gap-2 text-sm">
        <span className="font-mono text-foreground" title={address}>
          {shortenAddress(address)}
        </span>
        <span className="rounded-md bg-border px-2 py-0.5 text-xs text-muted">{statusLabel}</span>
      </span>
      <button
        type="button"
        onClick={() => disconnect.mutate()}
        className="text-xs text-muted underline-offset-4 hover:text-foreground hover:underline"
      >
        Disconnect
      </button>
    </div>
  );
}
