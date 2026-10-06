import { Check } from "lucide-react";

import { Spinner } from "@/components/ui/spinner";
import {
  walletPaymentStepLabels,
  walletPaymentStepOrder,
  walletPaymentStepState,
  type WalletPaymentStep,
} from "@/features/chapter-purchase/wallet-payment-steps";
import { cn } from "@/utils/class-names";

export function WalletPaymentProgress({ currentStep }: { currentStep: WalletPaymentStep }) {
  return (
    <div className="rounded-xl border border-border bg-background/60 p-4">
      <p className="text-sm text-muted">Confirm each request in your wallet. Keep this window open.</p>
      <ol className="mt-3 space-y-2.5">
        {walletPaymentStepOrder.map((step) => {
          const stepState = walletPaymentStepState(step, currentStep);
          return (
            <li key={step} className="flex items-center gap-3 text-sm">
              <span
                className={cn(
                  "flex size-6 shrink-0 items-center justify-center rounded-full border",
                  stepState === "DONE" && "border-up/50 bg-up/15 text-up",
                  stepState === "ACTIVE" && "border-accent text-accent",
                  stepState === "WAITING" && "border-border text-subtle",
                )}
              >
                {stepState === "DONE" ? (
                  <Check className="size-3.5" aria-hidden="true" />
                ) : stepState === "ACTIVE" ? (
                  <Spinner className="size-3.5" label={walletPaymentStepLabels[step]} />
                ) : null}
              </span>
              <span className={stepState === "WAITING" ? "text-subtle" : "text-foreground"}>
                {walletPaymentStepLabels[step]}
              </span>
            </li>
          );
        })}
      </ol>
    </div>
  );
}
