"use client";

import { Check, Copy, CreditCard, Wallet } from "lucide-react";

import {
  ethDecimals,
  shownEthDecimals,
  shownUsdcDecimals,
  usdcDecimals,
} from "@/features/chapter-purchase/wallet-payment-asset";
import type { Subscription, WalletPaymentDetails } from "@/features/subscriptions/api/subscription-api";
import { useCopyToClipboard } from "@/utils/clipboard/use-copy-to-clipboard";
import { formatTokenAmount } from "@/utils/web3/format-token-amount";
import { shortenAddress } from "@/utils/web3/shorten-address";

function formatPaidAmount(walletPayment: WalletPaymentDetails): string {
  const isEth = walletPayment.asset === "ETH";
  const amount = formatTokenAmount(
    BigInt(walletPayment.amount_units),
    isEth ? ethDecimals : usdcDecimals,
    isEth ? shownEthDecimals : shownUsdcDecimals,
  );
  return `${amount} ${walletPayment.asset}`;
}

function WalletPaymentLine({ walletPayment }: { walletPayment: WalletPaymentDetails }) {
  const { copyText, hasCopied } = useCopyToClipboard();
  return (
    <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted">
      <span className="flex items-center gap-2">
        <Wallet className="size-4 text-accent" aria-hidden="true" />
        Paid <span className="font-mono text-foreground">{formatPaidAmount(walletPayment)}</span> from{" "}
        <span className="font-mono text-foreground" title={walletPayment.payer_address}>
          {shortenAddress(walletPayment.payer_address)}
        </span>
      </span>
      <button
        type="button"
        onClick={() => void copyText(walletPayment.transaction_hash)}
        title={walletPayment.transaction_hash}
        className="flex items-center gap-1.5 rounded-md px-1.5 py-0.5 font-mono text-xs transition-colors hover:bg-border hover:text-foreground"
      >
        Transaction {shortenAddress(walletPayment.transaction_hash)}
        {hasCopied ? (
          <Check className="size-3.5 text-up" aria-label="Copied" />
        ) : (
          <Copy className="size-3.5" aria-label="Copy transaction hash" />
        )}
      </button>
    </div>
  );
}

export function SubscriptionPaymentDetails({ subscription }: { subscription: Subscription }) {
  if (subscription.payment_method === "WALLET") {
    return subscription.wallet_payment ? (
      <WalletPaymentLine walletPayment={subscription.wallet_payment} />
    ) : (
      <p className="flex items-center gap-2 text-sm text-muted">
        <Wallet className="size-4 text-accent" aria-hidden="true" />
        Paid from a wallet
      </p>
    );
  }
  return (
    <p className="flex items-center gap-2 text-sm text-muted">
      <CreditCard className="size-4" aria-hidden="true" />
      Paid by card
    </p>
  );
}
