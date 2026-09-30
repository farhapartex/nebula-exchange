"use client";

import { CreditCard, Coins, Wallet, type LucideIcon } from "lucide-react";
import { RadioGroup } from "radix-ui";

import { NcAmount } from "@/components/money/nc-amount";
import type { PaymentChoice, PaymentChoiceAvailability } from "@/features/payments/payment-method-rules";
import { cn } from "@/utils/class-names";

const choicePresentation: Record<PaymentChoice, { label: string; description: string; icon: LucideIcon }> = {
  balance: { label: "NC balance", description: "Instant, no fees.", icon: Coins },
  card: { label: "Card", description: "Secure checkout by Stripe.", icon: CreditCard },
  wallet: { label: "USDC on Base", description: "Pay from your linked wallet.", icon: Wallet },
};

type PaymentMethodPickerProps = {
  choices: PaymentChoiceAvailability[];
  value: PaymentChoice | null;
  onValueChange: (choice: PaymentChoice) => void;
  availableBalance?: string | null;
  label?: string;
};

export function PaymentMethodPicker({
  choices,
  value,
  onValueChange,
  availableBalance,
  label = "Pay with",
}: PaymentMethodPickerProps) {
  return (
    <fieldset>
      <legend className="mb-3 text-sm font-medium text-foreground">{label}</legend>
      <RadioGroup.Root
        value={value ?? undefined}
        onValueChange={(nextChoice) => onValueChange(nextChoice as PaymentChoice)}
        className={cn("grid gap-3", choices.length === 3 ? "sm:grid-cols-3" : "sm:grid-cols-2")}
      >
        {choices.map((availability) => {
          const presentation = choicePresentation[availability.choice];
          const ChoiceIcon = presentation.icon;
          return (
            <RadioGroup.Item
              key={availability.choice}
              value={availability.choice}
              disabled={!availability.isAllowed}
              className={cn(
                "flex items-start gap-3 rounded-xl border p-3 text-left transition-colors",
                "focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none",
                "border-border bg-background/40 hover:border-border-strong",
                "data-[state=checked]:border-accent/60 data-[state=checked]:bg-accent/10 data-[state=checked]:shadow-[0_0_20px_-10px] data-[state=checked]:shadow-accent",
                "disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:border-border",
              )}
            >
              <ChoiceIcon
                className={cn(
                  "mt-0.5 size-5 shrink-0",
                  value === availability.choice ? "text-accent-soft" : "text-muted",
                )}
                aria-hidden="true"
              />
              <span className="min-w-0">
                <span className="block text-sm font-medium text-foreground">{presentation.label}</span>
                <span className="block text-xs text-muted">
                  {availability.unavailableReason ??
                    (availability.choice === "balance" && availableBalance ? (
                      <>
                        <NcAmount amount={availableBalance} /> available
                      </>
                    ) : (
                      presentation.description
                    ))}
                </span>
              </span>
            </RadioGroup.Item>
          );
        })}
      </RadioGroup.Root>
    </fieldset>
  );
}
