"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";

import { createPayment, type CreatePaymentRequest } from "@/features/payments/api/payments-api";
import { createIdempotencyKey } from "@/lib/api/api-client";

export function useCardCheckout(onError?: (error: Error) => void) {
  const [isRedirecting, setIsRedirecting] = useState(false);
  const checkoutMutation = useMutation({
    mutationFn: (createRequest: Omit<CreatePaymentRequest, "method">) =>
      createPayment({ ...createRequest, method: "card" }, createIdempotencyKey()),
    meta: { showsThrottlingInline: true },
    onSuccess: (payment) => {
      if (payment.checkout_url) {
        setIsRedirecting(true);
        window.location.assign(payment.checkout_url);
      }
    },
    onError,
  });
  return {
    startCheckout: checkoutMutation.mutate,
    isStartingCheckout: checkoutMutation.isPending || isRedirecting,
    checkoutError: checkoutMutation.error,
  };
}
