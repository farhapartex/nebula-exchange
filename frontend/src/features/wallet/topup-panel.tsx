"use client";

import { useState } from "react";
import { CircleAlert, Lock } from "lucide-react";
import { RadioGroup } from "radix-ui";

import { NcAmount } from "@/components/money/nc-amount";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { PaymentMethodPicker } from "@/features/payments/components/payment-method-picker";
import { describePaymentChoices } from "@/features/payments/payment-method-rules";
import { useCardCheckout } from "@/features/payments/use-card-checkout";
import { toMicroUnits, topupAmountsInNc, type TopupAmount } from "@/features/wallet/topup-amounts";
import { isApiError } from "@/lib/api/api-error";
import { cn } from "@/utils/class-names";

const topupChoices = describePaymentChoices("topup", 0n, null);

function limitDetails(error: Error | null): { window: string; remaining: string } | null {
  if (!isApiError(error) || error.code !== "LIMIT_EXCEEDED") {
    return null;
  }
  const details = (error.details ?? {}) as { window?: string; remaining?: string };
  return { window: details.window ?? "daily", remaining: details.remaining ?? "0" };
}

export function TopUpPanel() {
  const [selectedAmount, setSelectedAmount] = useState<TopupAmount>("10");
  const { startCheckout, isStartingCheckout, checkoutError } = useCardCheckout();
  const reachedLimit = limitDetails(checkoutError);
  const otherError = checkoutError && !reachedLimit ? checkoutError.message : null;

  return (
    <Card>
      <CardHeader title="Top up NC" description="Card NC is spendable everywhere but can never be withdrawn." />
      <CardContent className="space-y-5">
        <fieldset>
          <legend className="mb-3 text-sm font-medium text-foreground">Amount</legend>
          <RadioGroup.Root
            value={selectedAmount}
            onValueChange={(nextAmount) => setSelectedAmount(nextAmount as TopupAmount)}
            className="grid grid-cols-3 gap-2 sm:grid-cols-5"
          >
            {topupAmountsInNc.map((amountInNc) => (
              <RadioGroup.Item
                key={amountInNc}
                value={amountInNc}
                className={cn(
                  "h-12 rounded-lg border border-border bg-background/40 font-mono text-sm text-foreground tabular-nums transition-colors",
                  "hover:border-border-strong focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none",
                  "data-[state=checked]:border-accent/60 data-[state=checked]:bg-accent/10",
                )}
              >
                {amountInNc} NC
              </RadioGroup.Item>
            ))}
          </RadioGroup.Root>
        </fieldset>

        <PaymentMethodPicker choices={topupChoices} value="card" onValueChange={() => undefined} />

        {reachedLimit && (
          <div
            role="alert"
            className="flex gap-3 rounded-lg border border-warning/40 bg-warning/10 px-3 py-2.5 text-sm"
          >
            <CircleAlert className="mt-0.5 size-4 shrink-0 text-warning" aria-hidden="true" />
            <p className="text-foreground">
              {reachedLimit.window === "daily"
                ? "Daily card limit reached (500 NC per day)."
                : "30-day card limit reached (2,000 NC)."}{" "}
              You can still add <NcAmount amount={reachedLimit.remaining} className="text-foreground" /> by card.
            </p>
          </div>
        )}
        {otherError && (
          <p
            role="alert"
            className="rounded-lg border border-down/40 bg-down-soft/40 px-3 py-2 text-sm text-foreground"
          >
            {otherError}
          </p>
        )}

        <Button
          size="lg"
          className="w-full"
          isLoading={isStartingCheckout}
          onClick={() => startCheckout({ purpose: "TOPUP", amount_nc: selectedAmount })}
        >
          <Lock className="size-4" aria-hidden="true" />
          Pay <NcAmount amount={toMicroUnits(selectedAmount)} /> by card
        </Button>
        <p className="text-center text-xs text-subtle">Limits: 500 NC per day and 2,000 NC per 30 days by card.</p>
      </CardContent>
    </Card>
  );
}
