"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Minus, Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { fetchPlans, plansQueryKey, type Plan } from "@/features/chapter-purchase/api/plan-api";
import { describePlanOption } from "@/features/chapter-purchase/plan-option-label";
import { PurchaseChoiceCard } from "@/features/chapter-purchase/purchase-choice-card";
import { PurchaseSummary } from "@/features/chapter-purchase/purchase-summary";
import { publishToastEvent } from "@/lib/notifications/toast-events";
import { formatUsd } from "@/utils/money/format-usd";

const mockCheckoutDelayInMilliseconds = 900;

type ChapterPurchaseDialogProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
};

export function ChapterPurchaseDialog({ isOpen, onOpenChange }: ChapterPurchaseDialogProps) {
  const plansQuery = useQuery({ queryKey: plansQueryKey, queryFn: fetchPlans, staleTime: 0, enabled: isOpen });
  const availablePlans = (plansQuery.data ?? []).filter((plan) => plan.is_available && plan.options.length > 0);
  const [chosenPlanID, setChosenPlanID] = useState<string | null>(null);
  const [optionIndexByPlan, setOptionIndexByPlan] = useState<Record<string, number>>({});
  const [isPaying, setIsPaying] = useState(false);

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
    setIsPaying(true);
    window.setTimeout(() => {
      setIsPaying(false);
      onOpenChange(false);
      publishToastEvent({
        tone: "info",
        title: "Checkout is not connected yet",
        description: `This will open Stripe for the ${selectedPlan.name.toLowerCase()} plan: ${selectedOption.chapter_count} ${selectedOption.chapter_count === 1 ? "chapter" : "chapters"} for ${formatUsd(BigInt(selectedOption.total_cents))}.`,
      });
    }, mockCheckoutDelayInMilliseconds);
  }

  return (
    <Dialog
      isOpen={isOpen}
      onOpenChange={onOpenChange}
      title="Unlock the story"
      description="Pay once with your card. Unlocked chapters stay yours to play and replay."
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
          <fieldset className="space-y-3">
            <legend className="sr-only">Choose a plan</legend>
            {availablePlans.map((plan) => {
              const planOption = plan.options[optionIndexOf(plan)];
              const isBundle = plan.kind === "CHAPTER_BUNDLE";
              return (
                <PurchaseChoiceCard
                  key={plan.id}
                  planID={plan.id}
                  isSelected={plan.id === selectedPlan.id}
                  onSelect={setChosenPlanID}
                  title={plan.name}
                  detail={describePlanOption(plan, planOption)}
                  price={plan.kind === "SINGLE_CHAPTER" ? formatUsd(BigInt(planOption.total_cents)) : undefined}
                  discountPercent={planOption.discount_percent}
                >
                  {isBundle && plan.options.length > 1 && (
                    <div className="mt-3 flex items-center gap-3">
                      <Button
                        type="button"
                        size="sm"
                        variant="secondary"
                        aria-label="Fewer chapters"
                        disabled={optionIndexOf(plan) === 0}
                        onClick={() => changeOption(plan, -1)}
                      >
                        <Minus className="size-3.5" />
                      </Button>
                      <span className="w-24 text-center font-mono text-sm whitespace-nowrap text-foreground tabular-nums">
                        {planOption.chapter_count} chapters
                      </span>
                      <Button
                        type="button"
                        size="sm"
                        variant="secondary"
                        aria-label="More chapters"
                        disabled={optionIndexOf(plan) === plan.options.length - 1}
                        onClick={() => changeOption(plan, 1)}
                      >
                        <Plus className="size-3.5" />
                      </Button>
                    </div>
                  )}
                </PurchaseChoiceCard>
              );
            })}
          </fieldset>

          <PurchaseSummary option={selectedOption} />

          <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
            <Button type="button" variant="secondary" size="lg" onClick={() => onOpenChange(false)} disabled={isPaying}>
              Cancel
            </Button>
            <Button type="button" size="lg" onClick={pay} isLoading={isPaying} className="sm:min-w-52">
              {isPaying ? "Opening checkout" : `Pay ${formatUsd(BigInt(selectedOption.total_cents))}`}
            </Button>
          </div>
        </div>
      )}
    </Dialog>
  );
}
