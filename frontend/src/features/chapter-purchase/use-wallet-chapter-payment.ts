"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { useRouter } from "next/navigation";

import { sendWalletPayment } from "@/features/chapter-purchase/send-wallet-payment";
import { walletPaymentErrorToast } from "@/features/chapter-purchase/wallet-payment-errors";
import type { WalletPaymentAsset } from "@/features/chapter-purchase/wallet-payment-asset";
import type { WalletPaymentStep } from "@/features/chapter-purchase/wallet-payment-steps";
import {
  createWalletCheckout,
  reportWalletTransaction,
  type CheckoutSessionRequest,
} from "@/features/subscriptions/api/checkout-session-api";
import { checkoutSessionQueryParameter } from "@/features/subscriptions/checkout-return-query";
import { publishToastEvent } from "@/lib/notifications/toast-events";

export type WalletChapterPaymentInput = {
  checkoutRequest: CheckoutSessionRequest;
  asset: WalletPaymentAsset;
  payerAddress: `0x${string}`;
};

export function useWalletChapterPayment() {
  const router = useRouter();
  const [walletPaymentStep, setWalletPaymentStep] = useState<WalletPaymentStep>("PREPARING");

  const walletPaymentMutation = useMutation({
    mutationFn: async (paymentInput: WalletChapterPaymentInput) => {
      setWalletPaymentStep("PREPARING");
      const walletCheckout = await createWalletCheckout(paymentInput.checkoutRequest);
      const transactionHash = await sendWalletPayment(
        {
          walletCheckout,
          asset: paymentInput.asset,
          payerAddress: paymentInput.payerAddress,
        },
        setWalletPaymentStep,
      );
      await reportWalletTransaction(walletCheckout.id, transactionHash).catch(() => null);
      return walletCheckout;
    },
    onSuccess: (walletCheckout) => {
      const subscriptionQuery = new URLSearchParams({ [checkoutSessionQueryParameter]: walletCheckout.id });
      router.push(`/subscription?${subscriptionQuery.toString()}`);
    },
    onError: (error) => publishToastEvent(walletPaymentErrorToast(error)),
  });

  return { walletPaymentMutation, walletPaymentStep };
}
