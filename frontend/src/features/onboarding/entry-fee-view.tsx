"use client";

import { useRouter } from "next/navigation";
import { Lock } from "lucide-react";

import { PageContainer } from "@/components/layout/page-container";
import { NcAmount } from "@/components/money/nc-amount";
import { Button } from "@/components/ui/button";
import { fetchCurrentUser } from "@/features/auth/api/session-api";
import { useAuth } from "@/features/auth/session/use-auth";
import { entryFeeInMicroUnits } from "@/features/onboarding/starter-pack";
import { StarterPackPreview } from "@/features/onboarding/starter-pack-preview";
import { PaymentMethodPicker } from "@/features/payments/components/payment-method-picker";
import { describePaymentChoices } from "@/features/payments/payment-method-rules";
import { useCardCheckout } from "@/features/payments/use-card-checkout";
import { isApiError } from "@/lib/api/api-error";

const entryFeeChoices = describePaymentChoices("entry_fee", BigInt(entryFeeInMicroUnits), 0n);

export function EntryFeeView() {
  const router = useRouter();
  const { replaceUser } = useAuth();
  const { startCheckout, isStartingCheckout, checkoutError } = useCardCheckout(async (error) => {
    if (isApiError(error) && error.code === "CONFLICT") {
      replaceUser(await fetchCurrentUser());
      router.replace("/hangar");
    }
  });

  const errorMessage =
    checkoutError && isApiError(checkoutError) && checkoutError.code !== "CONFLICT"
      ? checkoutError.message
      : checkoutError
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
        <PaymentMethodPicker choices={entryFeeChoices} value="card" onValueChange={() => undefined} />

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
          isLoading={isStartingCheckout}
          onClick={() => startCheckout({ purpose: "ENTRY_FEE" })}
        >
          <Lock className="size-4" aria-hidden="true" />
          Pay <NcAmount amount={entryFeeInMicroUnits} /> by card
        </Button>
        <p className="mt-2 text-center text-xs text-subtle">You are charged the same amount in USD. 1 NC = 1 USD.</p>
      </div>
    </PageContainer>
  );
}
