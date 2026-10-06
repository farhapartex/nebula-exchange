"use client";

import Image from "next/image";
import { Wallet } from "lucide-react";
import { useConnect, useConnectors, type Connector } from "wagmi";

import { Spinner } from "@/components/ui/spinner";
import { publishToastEvent } from "@/lib/notifications/toast-events";

const genericInjectedConnectorID = "injected";

function walletsToOffer(connectors: readonly Connector[]): readonly Connector[] {
  const discoveredWallets = connectors.filter((connector) => connector.id !== genericInjectedConnectorID);
  return discoveredWallets.length > 0 ? discoveredWallets : connectors;
}

export function WalletConnectorList() {
  const connectors = useConnectors();
  const connect = useConnect();
  const offeredWallets = walletsToOffer(connectors);
  const hasBrowserWallet = typeof window !== "undefined" && (offeredWallets.length > 1 || "ethereum" in window);

  if (!hasBrowserWallet) {
    return (
      <p className="rounded-lg border border-border bg-background/50 p-3 text-sm text-muted">
        No browser wallet found. Install MetaMask or Coinbase Wallet, then reload this page.
      </p>
    );
  }

  function connectWallet(connector: Connector) {
    connect.mutate(
      { connector },
      {
        onError: () =>
          publishToastEvent({
            tone: "info",
            title: "Wallet not connected",
            description: "The wallet did not connect. Open it and try again.",
          }),
      },
    );
  }

  return (
    <ul className="grid gap-2">
      {offeredWallets.map((connector) => {
        const isConnecting = connect.isPending && connect.variables?.connector === connector;
        return (
          <li key={connector.uid}>
            <button
              type="button"
              onClick={() => connectWallet(connector)}
              disabled={connect.isPending}
              className="flex w-full items-center gap-3 rounded-lg border border-border bg-background/50 px-3 py-2.5 text-left text-sm text-foreground transition-colors hover:border-border-strong focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none disabled:opacity-60"
            >
              {connector.icon ? (
                <Image src={connector.icon} alt="" width={24} height={24} unoptimized className="size-6 rounded" />
              ) : (
                <Wallet className="size-5 text-muted" aria-hidden="true" />
              )}
              <span className="flex-1 font-medium">
                {connector.id === genericInjectedConnectorID ? "Browser wallet" : connector.name}
              </span>
              {isConnecting && <Spinner label={`Connecting ${connector.name}`} />}
            </button>
          </li>
        );
      })}
    </ul>
  );
}
