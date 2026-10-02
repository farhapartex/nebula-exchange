"use client";

import { useEffect, useRef } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useMutation } from "@tanstack/react-query";
import { CircleSlash } from "lucide-react";

import { PageContainer } from "@/components/layout/page-container";
import { Button } from "@/components/ui/button";
import { cancelPayment } from "@/features/payments/api/payments-api";
import { purposeCopy } from "@/features/payments/payment-purpose-copy";
import { isApiError } from "@/lib/api/api-error";
import { useRedirectCountdown } from "@/utils/navigation/use-redirect-countdown";

const redirectDelayInSeconds = 3;

export function PaymentCancelledView() {
  const searchParameters = useSearchParams();
  const paymentID = searchParameters.get("id") ?? "";
  const cancelMutation = useMutation({ mutationFn: cancelPayment });
  const hasRequestedCancellation = useRef(false);

  useEffect(() => {
    if (paymentID && !hasRequestedCancellation.current) {
      hasRequestedCancellation.current = true;
      cancelMutation.mutate(paymentID);
    }
  }, [cancelMutation, paymentID]);

  const wasAlreadyPaid = isApiError(cancelMutation.error) && cancelMutation.error.code === "CONFLICT";
  const returnPath = cancelMutation.data ? purposeCopy[cancelMutation.data.purpose].retryHref : "/hangar";
  const remainingSeconds = useRedirectCountdown(returnPath, redirectDelayInSeconds, cancelMutation.isSuccess);

  return (
    <PageContainer className="max-w-lg py-12 sm:py-16">
      <div
        aria-live="polite"
        className="flex flex-col items-center gap-3 rounded-2xl border border-border bg-surface/80 px-6 py-10 text-center"
      >
        <CircleSlash className="size-10 text-muted" aria-hidden="true" />
        {wasAlreadyPaid ? (
          <>
            <h1 className="text-xl font-semibold text-foreground">This payment already went through</h1>
            <p className="max-w-sm text-sm text-muted">It was completed before you cancelled.</p>
            <Button asChild>
              <Link href={`/payment/result?id=${paymentID}`}>See the result</Link>
            </Button>
          </>
        ) : (
          <>
            <h1 className="text-xl font-semibold text-foreground">Payment cancelled</h1>
            <p className="max-w-sm text-sm text-muted">
              Nothing was charged.{" "}
              {cancelMutation.isSuccess ? `Taking you back in ${remainingSeconds}s…` : "Closing the checkout…"}
            </p>
            {cancelMutation.isError && (
              <Button asChild variant="secondary">
                <Link href="/hangar">Back to the hangar</Link>
              </Button>
            )}
          </>
        )}
      </div>
    </PageContainer>
  );
}
