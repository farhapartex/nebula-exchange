import type { PlanOption } from "@/features/chapter-purchase/api/plan-api";
import { formatUsd } from "@/utils/money/format-usd";

export function PurchaseSummary({ option }: { option: PlanOption }) {
  const discountCents = BigInt(option.discount_cents);
  return (
    <div className="rounded-xl border border-border bg-background/60 p-4">
      <p className="text-xs font-semibold tracking-[0.18em] text-subtle uppercase">You will be charged</p>
      <ul className="mt-3 space-y-1.5 text-sm">
        {option.chapters.map((chapter) => (
          <li key={chapter.id} className="flex justify-between gap-4">
            <span className="text-muted">
              Chapter {chapter.number} · {chapter.title}
            </span>
            <span className="font-mono whitespace-nowrap text-foreground tabular-nums">
              {formatUsd(BigInt(chapter.price_cents))}
            </span>
          </li>
        ))}
      </ul>
      <dl className="mt-3 space-y-1.5 border-t border-border pt-3 text-sm">
        <div className="flex justify-between gap-4">
          <dt className="text-muted">Subtotal</dt>
          <dd className="font-mono whitespace-nowrap text-foreground tabular-nums">
            {formatUsd(BigInt(option.subtotal_cents))}
          </dd>
        </div>
        {discountCents > 0n && (
          <div className="flex justify-between gap-4">
            <dt className="text-up">Discount ({option.discount_percent}%)</dt>
            <dd className="font-mono whitespace-nowrap text-up tabular-nums">{formatUsd(-discountCents)}</dd>
          </div>
        )}
        <div className="flex justify-between gap-4 border-t border-border pt-2 text-base">
          <dt className="font-semibold text-foreground">Total</dt>
          <dd className="font-mono font-semibold whitespace-nowrap text-foreground tabular-nums">
            {formatUsd(BigInt(option.total_cents))}
          </dd>
        </div>
      </dl>
    </div>
  );
}
