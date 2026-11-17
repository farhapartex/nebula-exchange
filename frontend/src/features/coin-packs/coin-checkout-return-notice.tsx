"use client";

import { useEffect } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

import { Spinner } from "@/components/ui/spinner";
import type { ToastRequest } from "@/components/ui/toast/toast-context";
import {
  coinCheckoutCancelledValue,
  coinCheckoutQueryParameter,
} from "@/features/coin-packs/coin-checkout-return-query";
import {
  useCheckoutConfirmation,
  type CheckoutConfirmationState,
} from "@/features/subscriptions/use-checkout-confirmation";
import { publishToastEvent } from "@/lib/notifications/toast-events";

const finalToastByState: Partial<Record<CheckoutConfirmationState, ToastRequest>> = {
  PAID: { tone: "success", title: "Coins added", description: "Your new coins are ready to spend." },
  EXPIRED: { tone: "info", title: "Payment not completed", description: "No money was taken and no coins were added." },
  DELAYED: {
    tone: "info",
    title: "Still waiting for Stripe",
    description: "Your coins are added as soon as the payment arrives.",
  },
  FAILED: {
    tone: "error",
    title: "We could not check this payment",
    description: "If you were charged, your coins are added automatically.",
  },
};

const cancelledToast: ToastRequest = {
  tone: "info",
  title: "Payment cancelled",
  description: "No money was taken. You can get coins any time.",
};

export function CoinCheckoutReturnNotice() {
  const router = useRouter();
  const pathname = usePathname();
  const returnedValue = useSearchParams().get(coinCheckoutQueryParameter);
  const isCancelled = returnedValue === coinCheckoutCancelledValue;
  const checkoutSessionID = returnedValue && !isCancelled ? returnedValue : null;
  const confirmation = useCheckoutConfirmation(checkoutSessionID);
  const confirmationState = confirmation?.state ?? null;

  useEffect(() => {
    if (isCancelled) {
      publishToastEvent(cancelledToast);
      router.replace(pathname);
      return;
    }
    const finalToast = confirmationState ? finalToastByState[confirmationState] : undefined;
    if (!finalToast) {
      return;
    }
    publishToastEvent(finalToast);
    router.replace(pathname);
  }, [confirmationState, isCancelled, pathname, router]);

  if (confirmationState !== "CONFIRMING") {
    return null;
  }
  return (
    <div
      role="status"
      className="fixed right-4 bottom-4 z-40 flex items-center gap-3 rounded-xl border border-border-strong bg-surface-raised px-4 py-3 text-sm text-foreground shadow-xl shadow-black/40"
    >
      <Spinner className="size-4" label="Confirming payment" />
      Adding your coins…
    </div>
  );
}
