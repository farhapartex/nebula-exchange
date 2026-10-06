"use client";

import { useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";

import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { fetchPlans, plansQueryKey, type Plan } from "@/features/chapter-purchase/api/plan-api";
import { PaymentMethodPicker } from "@/features/chapter-purchase/payment-method-picker";
import { PlanChoiceList } from "@/features/chapter-purchase/plan-choice-list";
import { PurchaseSummary } from "@/features/chapter-purchase/purchase-summary";
import { useWalletChapterPayment } from "@/features/chapter-purchase/use-wallet-chapter-payment";
import { WalletPaymentProgress } from "@/features/chapter-purchase/wallet-payment-progress";
import { createCheckoutSession, type PaymentMethod } from "@/features/subscriptions/api/checkout-session-api";
import { useWalletReadiness } from "@/features/wallet/use-wallet-readiness";
import { WalletReadinessPanel } from "@/features/wallet/wallet-readiness-panel";
import { publishToastEvent } from "@/lib/notifications/toast-events";
import { formatUsd } from "@/utils/money/format-usd";
import { formatUsdcFromCents } from "@/utils/money/format-usdc";

type ChapterPurchaseDialogProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
};

export function ChapterPurchaseDialog({ isOpen, onOpenChange }: ChapterPurchaseDialogProps) {
  const plansQuery = useQuery({ queryKey: plansQueryKey, queryFn: fetchPlans, staleTime: 0, enabled: isOpen });
  const availablePlans = (plansQuery.data ?? []).filter((plan) => plan.is_available && plan.options.length > 0);
  const [chosenPlanID, setChosenPlanID] = useState<string | null>(null);
  const [optionIndexByPlan, setOptionIndexByPlan] = useState<Record<string, number>>({});
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>("CARD");
  const isWalletPayment = paymentMethod === "WALLET";
  const walletReadiness = useWalletReadiness(isOpen && isWalletPayment);
  const { walletPaymentMutation, walletPaymentStep } = useWalletChapterPayment();
  const cardCheckoutMutation = useMutation({
    mutationFn: createCheckoutSession,
    onSuccess: (checkoutSession) => window.location.assign(checkoutSession.checkout_url),
    onError: () =>
      publishToastEvent({
        tone: "error",
        title: "Checkout could not start",
        description: "Nothing was charged. Try again in a moment.",
      }),
  });
  const isPayingByCard = cardCheckoutMutation.isPending || cardCheckoutMutation.isSuccess;
  const isPayingByWallet = walletPaymentMutation.isPending || walletPaymentMutation.isSuccess;
  const isPaying = isPayingByCard || isPayingByWallet;

  const selectedPlan = availablePlans.find((plan) => plan.id === chosenPlanID) ?? availablePlans[0];
  const optionIndexOf = (plan: Plan) => Math.min(optionIndexByPlan[plan.id] ?? 0, plan.options.length - 1);
  const selectedOption = selectedPlan?.options[optionIndexOf(selectedPlan)];

  function changeOption(plan: Plan, step: number) {
    setChosenPlanID(plan.id);
    setOptionIndexByPlan((indexes) => ({
      ...indexes,
      [plan.id]: Math.max(0, Math.min(plan.options.length - 1, optionIndexOf(plan) + step)),
    }));
  }

  function pay() {
    if (!selectedPlan || !selectedOption) {
      return;
    }
    const checkoutRequest = { plan_id: selectedPlan.id, chapter_count: selectedOption.chapter_count };
    if (isWalletPayment) {
      walletPaymentMutation.mutate(checkoutRequest);
      return;
    }
    cardCheckoutMutation.mutate(checkoutRequest);
  }

  function payButtonLabel(totalCents: bigint): string {
    if (isPayingByCard) {
      return "Opening Stripe";
    }
    if (isPayingByWallet) {
      return "Waiting for wallet";
    }
    return isWalletPayment ? `Pay ${formatUsdcFromCents(totalCents)}` : `Pay ${formatUsd(totalCents)}`;
  }

  return (
    <Dialog
      isOpen={isOpen}
      onOpenChange={(nextIsOpen) => !isPayingByWallet && onOpenChange(nextIsOpen)}
      title="Unlock the story"
      description="Pay once by card or from your wallet. Unlocked chapters stay yours to play and replay."
    >
      {plansQuery.isError ? (
        <ErrorState
          message="The plans could not be loaded. Try again."
          onRetry={() => void plansQuery.refetch()}
          isRetrying={plansQuery.isFetching}
        />
      ) : !plansQuery.data ? (
        <div className="space-y-3">
          <Skeleton className="h-20 rounded-xl" />
          <Skeleton className="h-20 rounded-xl" />
          <Skeleton className="h-32 rounded-xl" />
        </div>
      ) : !selectedPlan || !selectedOption ? (
        <p className="rounded-xl border border-border bg-background/60 p-4 text-sm text-muted">
          You already own every chapter that is out now. New chapters are on the way.
        </p>
      ) : (
        <div className="space-y-6">
          <PlanChoiceList
            plans={availablePlans}
            selectedPlanID={selectedPlan.id}
            optionIndexOf={optionIndexOf}
            onSelectPlan={setChosenPlanID}
            onChangeOption={changeOption}
            isDisabled={isPaying}
          />

          <PurchaseSummary option={selectedOption} />

          <PaymentMethodPicker selectedMethod={paymentMethod} onSelect={setPaymentMethod} isDisabled={isPaying} />

          {isWalletPayment &&
            (isPayingByWallet ? (
              <WalletPaymentProgress currentStep={walletPaymentStep} />
            ) : (
              <WalletReadinessPanel readiness={walletReadiness} />
            ))}

          <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
            <Button type="button" variant="secondary" size="lg" onClick={() => onOpenChange(false)} disabled={isPaying}>
              Cancel
            </Button>
            <Button
              type="button"
              size="lg"
              onClick={pay}
              isLoading={isPaying}
              disabled={isWalletPayment && walletReadiness.stage !== "READY"}
              className="sm:min-w-52"
            >
              {payButtonLabel(BigInt(selectedOption.total_cents))}
            </Button>
          </div>
        </div>
      )}
    </Dialog>
  );
}
