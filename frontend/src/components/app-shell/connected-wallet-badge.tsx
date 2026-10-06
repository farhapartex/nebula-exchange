"use client";

import { Wallet } from "lucide-react";
import { useConnection } from "wagmi";

import { shortenAddress } from "@/utils/web3/shorten-address";

export function ConnectedWalletBadge() {
  const connection = useConnection();
  if (connection.status !== "connected" || !connection.address) {
    return null;
  }
  return (
    <span
      title={`Wallet connected: ${connection.address}`}
      className="flex h-9 items-center gap-2 rounded-lg border border-border-strong bg-surface-raised px-3 text-sm"
    >
      <span className="relative flex">
        <Wallet className="size-4 text-accent" aria-hidden="true" />
        <span className="absolute -top-0.5 -right-0.5 size-2 rounded-full bg-up ring-2 ring-surface-raised" />
      </span>
      <span className="sr-only">Wallet connected</span>
      <span className="hidden font-mono text-foreground sm:inline">{shortenAddress(connection.address)}</span>
    </span>
  );
}
