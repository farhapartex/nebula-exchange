"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useMutation } from "@tanstack/react-query";
import { CreditCard, Lock, Wallet } from "lucide-react";

import { PageContainer } from "@/components/layout/page-container";
import { NcAmount } from "@/components/money/nc-amount";
import { Button } from "@/components/ui/button";
import { fetchCurrentUser } from "@/features/auth/api/session-api";
import { useAuth } from "@/features/auth/session/use-auth";
import { entryFeeInMicroUnits } from "@/features/onboarding/starter-pack";
import { StarterPackPreview } from "@/features/onboarding/starter-pack-preview";
import { createPayment } from "@/features/payments/api/payments-api";
import { createIdempotencyKey } from "@/lib/api/api-client";
import { isApiError } from "@/lib/api/api-error";
import { cn } from "@/utils/class-names";

export function EntryFeeView() {
  const router = useRouter();
  const { replaceUser } = useAuth();
  const [idempotencyKey] = useState(createIdempotencyKey);
  const [isRedirecting, setIsRedirecting] = useState(false);

  const payMutation = useMutation({
    mutationFn: () => createPayment({ purpose: "ENTRY_FEE", method: "card" }, idempotencyKey),
    meta: { showsThrottlingInline: true },
    onSuccess: (payment) => {
      if (payment.checkout_url) {
        setIsRedirecting(true);
        window.location.assign(payment.checkout_url);
      }
    },
    onError: async (error) => {
      if (isApiError(error) && error.code === "CONFLICT") {
        replaceUser(await fetchCurrentUser());
        router.replace("/hangar");
      }
    },
  });

  const errorMessage =
    payMutation.error && isApiError(payMutation.error) && payMutation.error.code !== "CONFLICT"
      ? payMutation.error.message
      : payMutation.error
        ? "We couldn't start the payment. Please try again."
        : null;

  return (
    <PageContainer className="max-w-3xl py-8 sm:py-10">
      <h1 className="text-2xl font-semibold tracking-tight text-foreground">Pay the entry fee</h1>
      <p className="mt-1 text-sm text-muted">
        One payment of <NcAmount amount={entryFeeInMicroUnits} className="text-foreground" /> unlocks the game and your
        starter pack.
      </p>

      <div className="mt-6">
        <StarterPackPreview />
      </div>

      <div className="mt-6 rounded-2xl border border-border bg-surface/70 p-5">
        <p className="mb-3 text-sm font-medium text-foreground">Pay with</p>
        <div className="grid gap-3 sm:grid-cols-2">
          <div
            className={cn(
              "flex items-start gap-3 rounded-xl border border-accent/60 bg-accent/10 p-3",
              "shadow-[0_0_20px_-10px] shadow-accent",
            )}
          >
            <CreditCard className="mt-0.5 size-5 text-accent-soft" aria-hidden="true" />
            <div>
              <p className="text-sm font-medium text-foreground">Card</p>
              <p className="text-xs text-muted">Secure checkout by Stripe.</p>
            </div>
          </div>
          <div className="flex items-start gap-3 rounded-xl border border-border bg-background/40 p-3 opacity-60">
            <Wallet className="mt-0.5 size-5 text-muted" aria-hidden="true" />
            <div>
              <p className="text-sm font-medium text-foreground">USDC on Base</p>
              <p className="text-xs text-muted">Link a wallet first. Coming with wallet support.</p>
            </div>
          </div>
        </div>

        {errorMessage && (
          <p
            role="alert"
            className="mt-4 rounded-lg border border-down/40 bg-down-soft/40 px-3 py-2 text-sm text-foreground"
          >
            {errorMessage}
          </p>
        )}

        <Button
          size="lg"
          className="mt-5 w-full"
          isLoading={payMutation.isPending || isRedirecting}
          onClick={() => payMutation.mutate()}
        >
          <Lock className="size-4" aria-hidden="true" />
          Pay <NcAmount amount={entryFeeInMicroUnits} /> by card
        </Button>
        <p className="mt-2 text-center text-xs text-subtle">You are charged the same amount in USD. 1 NC = 1 USD.</p>
      </div>
    </PageContainer>
  );
}
