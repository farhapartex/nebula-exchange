import type { SubscriptionStatus } from "@/features/subscriptions/api/subscription-api";
import { cn } from "@/utils/class-names";

const badgeClassNameByStatus: Record<SubscriptionStatus, string> = {
  PAID: "bg-up/15 text-up",
  REFUNDED: "bg-down-soft/40 text-down",
  DISPUTED: "bg-down-soft/40 text-down",
};

const badgeLabelByStatus: Record<SubscriptionStatus, string> = {
  PAID: "Paid",
  REFUNDED: "Refunded",
  DISPUTED: "Disputed",
};

export function SubscriptionStatusBadge({ status }: { status: SubscriptionStatus }) {
  return (
    <span className={cn("rounded-md px-2 py-0.5 text-xs font-semibold", badgeClassNameByStatus[status])}>
      {badgeLabelByStatus[status]}
    </span>
  );
}
