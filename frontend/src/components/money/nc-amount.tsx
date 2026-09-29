"use client";

import { Tooltip } from "@/components/ui/tooltip";
import { cn } from "@/utils/class-names";
import { formatMicroUnits, microUnitDecimals, type MicroUnitInput } from "@/utils/money/micro-units";

const defaultDisplayFractionDigits = 2;

type NcAmountProps = {
  amount: MicroUnitInput;
  fractionDigits?: number;
  showUnit?: boolean;
  showPlusSign?: boolean;
  className?: string;
};

export function NcAmount({
  amount,
  fractionDigits = defaultDisplayFractionDigits,
  showUnit = true,
  showPlusSign = false,
  className,
}: NcAmountProps) {
  const displayedAmount = formatMicroUnits(amount, { fractionDigits, showPlusSign });
  const exactAmount = formatMicroUnits(amount, { fractionDigits: microUnitDecimals, showPlusSign });
  const amountText = showUnit ? `${displayedAmount} NC` : displayedAmount;
  const isShowingFullPrecision = fractionDigits === microUnitDecimals;

  const amountElement = (
    <span
      tabIndex={isShowingFullPrecision ? undefined : 0}
      aria-label={`${exactAmount} NC`}
      className={cn("font-mono whitespace-nowrap tabular-nums", className)}
    >
      {amountText}
    </span>
  );

  if (isShowingFullPrecision) {
    return amountElement;
  }
  return <Tooltip content={<span className="font-mono tabular-nums">{exactAmount} NC</span>}>{amountElement}</Tooltip>;
}
