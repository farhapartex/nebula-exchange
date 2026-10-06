import { CreditCard, Wallet } from "lucide-react";
import type { ReactNode } from "react";

import type { PaymentMethod } from "@/features/subscriptions/api/checkout-session-api";
import { cn } from "@/utils/class-names";

type PaymentMethodOption = {
  method: PaymentMethod;
  label: string;
  detail: string;
  icon: ReactNode;
};

const paymentMethodOptions: PaymentMethodOption[] = [
  {
    method: "CARD",
    label: "Card",
    detail: "Pay in USD with Stripe",
    icon: <CreditCard className="size-5" aria-hidden="true" />,
  },
  {
    method: "WALLET",
    label: "Wallet",
    detail: "Pay in USDC with MetaMask, Coinbase Wallet and more",
    icon: <Wallet className="size-5" aria-hidden="true" />,
  },
];

type PaymentMethodPickerProps = {
  selectedMethod: PaymentMethod;
  onSelect: (paymentMethod: PaymentMethod) => void;
  isDisabled: boolean;
};

export function PaymentMethodPicker({ selectedMethod, onSelect, isDisabled }: PaymentMethodPickerProps) {
  return (
    <fieldset disabled={isDisabled}>
      <legend className="mb-2 text-xs font-semibold tracking-[0.18em] text-subtle uppercase">Pay with</legend>
      <div className="grid gap-3 sm:grid-cols-2">
        {paymentMethodOptions.map((option) => {
          const isSelected = option.method === selectedMethod;
          return (
            <label
              key={option.method}
              className={cn(
                "flex cursor-pointer items-start gap-3 rounded-xl border p-3.5 transition-colors has-[:disabled]:cursor-not-allowed has-[:focus-visible]:ring-2 has-[:focus-visible]:ring-highlight",
                isSelected ? "border-accent bg-accent/10" : "border-border bg-background/40 hover:border-border-strong",
              )}
            >
              <input
                type="radio"
                name="chapter-payment-method"
                value={option.method}
                checked={isSelected}
                onChange={() => onSelect(option.method)}
                className="sr-only"
              />
              <span className={isSelected ? "text-accent" : "text-muted"}>{option.icon}</span>
              <span>
                <span className="block font-medium text-foreground">{option.label}</span>
                <span className="mt-0.5 block text-xs text-muted">{option.detail}</span>
              </span>
            </label>
          );
        })}
      </div>
    </fieldset>
  );
}
