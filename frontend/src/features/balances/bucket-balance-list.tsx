import { NcAmount } from "@/components/money/nc-amount";
import type { BucketBalance } from "@/features/balances/api/balances-api";
import { bucketDescriptions } from "@/features/balances/bucket-descriptions";
import { cn } from "@/utils/class-names";

type BucketBalanceListProps = {
  buckets: BucketBalance[];
  className?: string;
};

export function BucketBalanceList({ buckets, className }: BucketBalanceListProps) {
  return (
    <ul className={cn("space-y-0.5", className)}>
      {buckets.map((bucketBalance) => {
        const bucketDescription = bucketDescriptions[bucketBalance.bucket];
        const isNegative = BigInt(bucketBalance.available) < 0n;
        return (
          <li key={bucketBalance.bucket} className="flex items-start justify-between gap-3 rounded-lg px-2.5 py-2">
            <div className="min-w-0">
              <p className="text-sm font-medium text-foreground">{bucketDescription.label}</p>
              <p className="text-xs text-muted">{bucketDescription.explanation}</p>
            </div>
            <div className="shrink-0 text-right text-sm">
              <NcAmount amount={bucketBalance.available} className={cn(isNegative ? "text-down" : "text-foreground")} />
              {BigInt(bucketBalance.held) > 0n && (
                <p className="text-xs text-muted">
                  + <NcAmount amount={bucketBalance.held} /> held
                </p>
              )}
            </div>
          </li>
        );
      })}
    </ul>
  );
}
