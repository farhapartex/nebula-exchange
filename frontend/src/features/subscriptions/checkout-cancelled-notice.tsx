"use client";

import { useEffect } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

import { checkoutCancelledValue, checkoutSessionQueryParameter } from "@/features/subscriptions/checkout-return-query";
import { publishToastEvent } from "@/lib/notifications/toast-events";

export function CheckoutCancelledNotice() {
  const router = useRouter();
  const pathname = usePathname();
  const isCheckoutCancelled = useSearchParams().get(checkoutSessionQueryParameter) === checkoutCancelledValue;

  useEffect(() => {
    if (!isCheckoutCancelled) {
      return;
    }
    publishToastEvent({
      tone: "info",
      title: "Payment cancelled",
      description: "No money was taken. You can unlock chapters any time.",
    });
    router.replace(pathname);
  }, [isCheckoutCancelled, pathname, router]);

  return null;
}
