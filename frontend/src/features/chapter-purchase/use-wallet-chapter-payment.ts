"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { UserRejectedRequestError } from "viem";

import { sendWalletPayment } from "@/features/chapter-purchase/send-wallet-payment";
import type { WalletPaymentStep } from "@/features/chapter-purchase/wallet-payment-steps";
import { createWalletCheckout, type CheckoutSessionRequest } from "@/features/subscriptions/api/checkout-session-api";
import { checkoutSessionQueryParameter } from "@/features/subscriptions/checkout-return-query";
import { publishToastEvent } from "@/lib/notifications/toast-events";

function isWalletRejection(error: unknown): boolean {
  return (
    error instanceof Error &&
    (error.name === UserRejectedRequestError.name || error.cause instanceof UserRejectedRequestError)
  );
}

export function useWalletChapterPayment() {
  const router = useRouter();
  const [walletPaymentStep, setWalletPaymentStep] = useState<WalletPaymentStep>("PREPARING");

  const walletPaymentMutation = useMutation({
    mutationFn: async (checkoutRequest: CheckoutSessionRequest) => {
      setWalletPaymentStep("PREPARING");
      const walletCheckout = await createWalletCheckout(checkoutRequest);
      await sendWalletPayment(walletCheckout, setWalletPaymentStep);
      return walletCheckout;
    },
    onSuccess: (walletCheckout) => {
      const subscriptionQuery = new URLSearchParams({ [checkoutSessionQueryParameter]: walletCheckout.id });
      router.push(`/subscription?${subscriptionQuery.toString()}`);
    },
    onError: (error) =>
      publishToastEvent(
        isWalletRejection(error)
          ? { tone: "info", title: "Payment cancelled", description: "You cancelled the request in your wallet." }
          : { tone: "error", title: "Wallet payment failed", description: "Nothing was taken. Try again in a moment." },
      ),
  });

  return { walletPaymentMutation, walletPaymentStep };
}
