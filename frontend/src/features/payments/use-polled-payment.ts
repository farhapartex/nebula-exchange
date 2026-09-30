"use client";

import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { fetchPayment, paymentQueryKey } from "@/features/payments/api/payments-api";
import { isPaymentSettled } from "@/features/payments/payment-outcome";

const pollIntervalInMilliseconds = 2_000;
const pollDurationInMilliseconds = 60_000;

export function usePolledPayment(paymentID: string) {
  const [pollingRound, setPollingRound] = useState(0);
  const [timedOutRoundKey, setTimedOutRoundKey] = useState<string | null>(null);
  const currentRoundKey = `${paymentID}:${pollingRound}`;
  const hasPollingTimedOut = timedOutRoundKey === currentRoundKey;

  useEffect(() => {
    const timeoutHandle = window.setTimeout(() => setTimedOutRoundKey(currentRoundKey), pollDurationInMilliseconds);
    return () => window.clearTimeout(timeoutHandle);
  }, [currentRoundKey]);

  const paymentQuery = useQuery({
    queryKey: paymentQueryKey(paymentID),
    queryFn: () => fetchPayment(paymentID),
    enabled: paymentID !== "",
    refetchInterval: (query) =>
      isPaymentSettled(query.state.data) || hasPollingTimedOut ? false : pollIntervalInMilliseconds,
  });

  return {
    paymentQuery,
    hasPollingTimedOut,
    restartPolling: () => {
      setPollingRound((currentRound) => currentRound + 1);
      void paymentQuery.refetch();
    },
  };
}
