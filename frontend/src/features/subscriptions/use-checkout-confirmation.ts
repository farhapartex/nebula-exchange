"use client";

import { useEffect, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import { currentPlayerQueryKey } from "@/features/auth/session/use-current-player";
import { plansQueryKey } from "@/features/chapter-purchase/api/plan-api";
import { fetchCheckoutSession } from "@/features/subscriptions/api/checkout-session-api";
import { subscriptionsQueryKey } from "@/features/subscriptions/api/subscription-api";

const confirmationPollIntervalInMilliseconds = 2000;
const confirmationWaitLimitInMilliseconds = 60000;

export type CheckoutConfirmationState = "CONFIRMING" | "PAID" | "EXPIRED" | "DELAYED" | "FAILED";

export function useCheckoutConfirmation(checkoutSessionID: string | null): CheckoutConfirmationState | null {
  const queryClient = useQueryClient();
  const [hasWaitedTooLong, setHasWaitedTooLong] = useState(false);

  const checkoutQuery = useQuery({
    queryKey: ["checkout-sessions", checkoutSessionID],
    queryFn: () => fetchCheckoutSession(checkoutSessionID ?? ""),
    enabled: checkoutSessionID !== null,
    retry: 1,
    refetchInterval: (query) =>
      query.state.data?.status === "OPEN" && !hasWaitedTooLong ? confirmationPollIntervalInMilliseconds : false,
  });
  const checkoutStatus = checkoutQuery.data?.status;

  useEffect(() => {
    if (checkoutSessionID === null) {
      return;
    }
    const waitLimitTimer = window.setTimeout(() => setHasWaitedTooLong(true), confirmationWaitLimitInMilliseconds);
    return () => window.clearTimeout(waitLimitTimer);
  }, [checkoutSessionID]);

  useEffect(() => {
    if (checkoutStatus !== "PAID") {
      return;
    }
    void queryClient.invalidateQueries({ queryKey: subscriptionsQueryKey });
    void queryClient.invalidateQueries({ queryKey: currentPlayerQueryKey });
    void queryClient.invalidateQueries({ queryKey: plansQueryKey });
  }, [checkoutStatus, queryClient]);

  if (checkoutSessionID === null) {
    return null;
  }
  if (checkoutQuery.isError) {
    return "FAILED";
  }
  if (checkoutStatus === "PAID" || checkoutStatus === "EXPIRED") {
    return checkoutStatus;
  }
  return hasWaitedTooLong ? "DELAYED" : "CONFIRMING";
}
