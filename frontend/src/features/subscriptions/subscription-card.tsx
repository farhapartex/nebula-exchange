import { HubPanel } from "@/features/fight-hub/hub-panel";
import type { Subscription } from "@/features/subscriptions/api/subscription-api";
import { SubscriptionStatusBadge } from "@/features/subscriptions/subscription-status-badge";
import { formatUsd } from "@/utils/money/format-usd";
import { formatCalendarDate } from "@/utils/time/format-calendar-date";

export function SubscriptionCard({ subscription }: { subscription: Subscription }) {
  const discountCents = BigInt(subscription.discount_cents);
  const isReversed = subscription.status !== "PAID";
  return (
    <HubPanel>
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="font-display text-xl tracking-[0.06em] text-foreground">{subscription.plan_name}</h2>
            <SubscriptionStatusBadge status={subscription.status} />
          </div>
          <p className="mt-1 text-sm text-muted">
            Paid {formatCalendarDate(subscription.paid_at)}
            {subscription.refunded_at &&
              ` · ${subscription.status === "DISPUTED" ? "Disputed" : "Refunded"} ${formatCalendarDate(subscription.refunded_at)}`}
          </p>
        </div>
        <div className="text-right">
          <p className="font-mono text-lg font-semibold whitespace-nowrap text-foreground tabular-nums">
            {formatUsd(BigInt(subscription.total_cents))}
          </p>
          {discountCents > 0n && (
            <p className="text-xs whitespace-nowrap text-up">
              {subscription.discount_percent}% off, saved {formatUsd(discountCents)}
            </p>
          )}
        </div>
      </div>
      <ul className="mt-4 flex flex-wrap gap-2">
        {subscription.chapters.map((chapter) => (
          <li
            key={chapter.id}
            className="rounded-lg border border-border bg-background/50 px-3 py-1.5 text-sm text-muted"
          >
            <span className="text-foreground">Chapter {chapter.number}</span> · {chapter.title}
          </li>
        ))}
      </ul>
      {isReversed && (
        <p className="mt-4 text-sm text-muted">
          These chapters are locked again. Your wins and stars in them are kept.
        </p>
      )}
    </HubPanel>
  );
}
