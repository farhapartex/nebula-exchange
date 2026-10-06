"use client";

import { Check, Copy } from "lucide-react";
import { DropdownMenu } from "radix-ui";
import { useConnection } from "wagmi";

import { publishToastEvent } from "@/lib/notifications/toast-events";
import { useCopyToClipboard } from "@/utils/clipboard/use-copy-to-clipboard";
import { shortenAddress } from "@/utils/web3/shorten-address";

export function ProfileMenuWallet() {
  const connection = useConnection();
  const { copyText, hasCopied } = useCopyToClipboard();
  if (connection.status !== "connected" || !connection.address) {
    return null;
  }
  const walletAddress = connection.address;

  async function copyWalletAddress(selectEvent: Event) {
    selectEvent.preventDefault();
    const isCopied = await copyText(walletAddress);
    if (!isCopied) {
      publishToastEvent({
        tone: "error",
        title: "Could not copy",
        description: "Copy the address from your wallet instead.",
      });
    }
  }

  return (
    <>
      <DropdownMenu.Separator className="my-1 h-px bg-border" />
      <p className="px-2.5 pt-1.5 text-[0.6875rem] font-semibold tracking-[0.18em] text-subtle uppercase">Wallet</p>
      <DropdownMenu.Item
        onSelect={(selectEvent) => void copyWalletAddress(selectEvent)}
        title={walletAddress}
        className="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-sm text-muted outline-none data-[highlighted]:bg-border data-[highlighted]:text-foreground"
      >
        <span className="size-2 shrink-0 rounded-full bg-up" aria-hidden="true" />
        <span className="flex-1 font-mono text-foreground">{shortenAddress(walletAddress)}</span>
        {hasCopied ? (
          <Check className="size-4 text-up" aria-label="Copied" />
        ) : (
          <Copy className="size-4" aria-label="Copy wallet address" />
        )}
      </DropdownMenu.Item>
    </>
  );
}
