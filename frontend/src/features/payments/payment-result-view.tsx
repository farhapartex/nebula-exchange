"use client";

import { useEffect, useRef, type ReactNode } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { CircleAlert, CircleCheck, Clock, Loader2, TriangleAlert } from "lucide-react";

import { PageContainer } from "@/components/layout/page-container";
import { NcAmount } from "@/components/money/nc-amount";
import { Button } from "@/components/ui/button";
import { fetchCurrentUser } from "@/features/auth/api/session-api";
import { useAuth } from "@/features/auth/session/use-auth";
import { balancesQueryKey } from "@/features/balances/api/balances-api";
import { inventoryQueryKey } from "@/features/inventory/api/inventory-api";
import { describePaymentOutcome, type PaymentOutcome } from "@/features/payments/payment-outcome";
import { purposeCopy, purposeFailureMessages } from "@/features/payments/payment-purpose-copy";
import { usePolledPayment } from "@/features/payments/use-polled-payment";
import { cn } from "@/utils/class-names";

const outcomeIcons: Record<PaymentOutcome, { icon: typeof CircleCheck; className: string }> = {
  processing: { icon: Loader2, className: "animate-spin text-accent-soft" },
  completed: { icon: CircleCheck, className: "text-up" },
  credited_only: { icon: TriangleAlert, className: "text-warning" },
  failed: { icon: CircleAlert, className: "text-down" },
  timed_out: { icon: Clock, className: "text-warning" },
};

export function PaymentResultView() {
  const searchParameters = useSearchParams();
  const paymentID = searchParameters.get("id") ?? "";
  const queryClient = useQueryClient();
  const { replaceUser } = useAuth();
  const { paymentQuery, hasPollingTimedOut, restartPolling } = usePolledPayment(paymentID);
  const payment = paymentQuery.data;
  const outcome = describePaymentOutcome(payment, hasPollingTimedOut);
  const hasRefreshedAccount = useRef(false);

  useEffect(() => {
    if ((outcome !== "completed" && outcome !== "credited_only") || hasRefreshedAccount.current) {
      return;
    }
    hasRefreshedAccount.current = true;
    void queryClient.invalidateQueries({ queryKey: balancesQueryKey });
    void queryClient.invalidateQueries({ queryKey: inventoryQueryKey });
    void fetchCurrentUser().then(replaceUser);
  }, [outcome, queryClient, replaceUser]);

  if (paymentID === "" || (paymentQuery.isError && !payment)) {
    return (
      <ResultCard
        outcome="failed"
        title="We can't find this payment"
        description="Check the link or start a new payment."
      >
        <Button asChild variant="secondary">
          <Link href="/hangar">Back to the hangar</Link>
        </Button>
      </ResultCard>
    );
  }

  const copy = purposeCopy[payment?.purpose ?? "TOPUP"];
  const { icon: OutcomeIcon, className: outcomeIconClassName } = outcomeIcons[outcome];

  return (
    <ResultCard
      outcome={outcome}
      icon={<OutcomeIcon className={cn("size-10", outcomeIconClassName)} aria-hidden="true" />}
      title={
        {
          processing: "Confirming your payment…",
          completed: copy.successTitle,
          credited_only: "Payment received",
          failed: "The payment didn't go through",
          timed_out: "Still waiting for confirmation",
        }[outcome]
      }
      description={
        {
          processing: "This usually takes a few seconds. You can keep this page open.",
          completed: copy.successDescription,
          credited_only:
            purposeFailureMessages[payment?.purpose_failure_code ?? ""] ?? "The NC stayed in your balance.",
          failed: "Nothing was charged. You can try again.",
          timed_out: "Your card provider is taking longer than usual. We'll notify you as soon as it's confirmed.",
        }[outcome]
      }
    >
      {payment && (
        <p className="text-sm text-muted">
          Amount <NcAmount amount={payment.amount} className="text-foreground" />
        </p>
      )}
      <div className="mt-2 flex flex-wrap justify-center gap-2">
        {outcome === "completed" && (
          <Button asChild>
            <Link href={copy.continueHref}>{copy.continueLabel}</Link>
          </Button>
        )}
        {outcome === "credited_only" && (
          <Button asChild>
            <Link href="/wallet">Open wallet</Link>
          </Button>
        )}
        {outcome === "failed" && (
          <Button asChild>
            <Link href={copy.retryHref}>Try again</Link>
          </Button>
        )}
        {outcome === "timed_out" && (
          <Button variant="secondary" onClick={restartPolling}>
            Check again
          </Button>
        )}
      </div>
      {process.env.NODE_ENV === "development" && payment?.status === "PENDING" && (
        <p className="mt-4 rounded-lg border border-dashed border-border-strong px-3 py-2 font-mono text-xs text-subtle">
          dev checkout: make dev-complete-payment id={payment.id}
        </p>
      )}
    </ResultCard>
  );
}

type ResultCardProps = {
  outcome: PaymentOutcome;
  icon?: ReactNode;
  title: string;
  description: string;
  children?: ReactNode;
};

function ResultCard({ outcome, icon, title, description, children }: ResultCardProps) {
  return (
    <PageContainer className="max-w-lg py-12 sm:py-16">
      <div
        aria-live="polite"
        data-outcome={outcome}
        className="flex flex-col items-center gap-3 rounded-2xl border border-border bg-surface/80 px-6 py-10 text-center"
      >
        {icon}
        <h1 className="text-xl font-semibold text-foreground">{title}</h1>
        <p className="max-w-sm text-sm text-muted">{description}</p>
        {children}
      </div>
    </PageContainer>
  );
}
