import { Minus, Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import type { Plan } from "@/features/chapter-purchase/api/plan-api";
import { describePlanOption } from "@/features/chapter-purchase/plan-option-label";
import { PurchaseChoiceCard } from "@/features/chapter-purchase/purchase-choice-card";
import { formatUsd } from "@/utils/money/format-usd";

type PlanChoiceListProps = {
  plans: Plan[];
  selectedPlanID: string;
  optionIndexOf: (plan: Plan) => number;
  onSelectPlan: (planID: string) => void;
  onChangeOption: (plan: Plan, step: number) => void;
  isDisabled: boolean;
};

export function PlanChoiceList({
  plans,
  selectedPlanID,
  optionIndexOf,
  onSelectPlan,
  onChangeOption,
  isDisabled,
}: PlanChoiceListProps) {
  return (
    <fieldset className="space-y-3" disabled={isDisabled}>
      <legend className="sr-only">Choose a plan</legend>
      {plans.map((plan) => {
        const planOption = plan.options[optionIndexOf(plan)];
        const isBundle = plan.kind === "CHAPTER_BUNDLE";
        return (
          <PurchaseChoiceCard
            key={plan.id}
            planID={plan.id}
            isSelected={plan.id === selectedPlanID}
            onSelect={onSelectPlan}
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
                  onClick={() => onChangeOption(plan, -1)}
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
                  onClick={() => onChangeOption(plan, 1)}
                >
                  <Plus className="size-3.5" />
                </Button>
              </div>
            )}
          </PurchaseChoiceCard>
        );
      })}
    </fieldset>
  );
}
