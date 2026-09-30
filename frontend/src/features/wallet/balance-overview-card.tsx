"use client";

import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { NcAmount } from "@/components/money/nc-amount";
import { BucketBalanceList } from "@/features/balances/bucket-balance-list";
import { useBalances } from "@/features/balances/use-balances";

export function BalanceOverviewCard() {
  const balancesQuery = useBalances();

  return (
    <Card>
      <CardHeader title="Balance" description="Your NC, split by where it came from." />
      <CardContent>
        {balancesQuery.isPending && <Skeleton className="h-64 rounded-xl" />}
        {balancesQuery.isError && (
          <ErrorState message="We couldn't load your balance." onRetry={() => void balancesQuery.refetch()} />
        )}
        {balancesQuery.data && (
          <div className="space-y-4">
            <div className="grid grid-cols-3 gap-3">
              {[
                { label: "Total", amount: balancesQuery.data.total, className: "text-foreground" },
                { label: "Held", amount: balancesQuery.data.held, className: "text-muted" },
                { label: "Withdrawable", amount: balancesQuery.data.withdrawable, className: "text-up" },
              ].map((summaryFigure) => (
                <div key={summaryFigure.label} className="rounded-xl border border-border bg-background/50 p-3">
                  <p className="text-xs text-muted">{summaryFigure.label}</p>
                  <p className="mt-1 text-lg font-semibold">
                    <NcAmount amount={summaryFigure.amount} className={summaryFigure.className} />
                  </p>
                </div>
              ))}
            </div>
            <BucketBalanceList buckets={balancesQuery.data.buckets} className="-mx-2.5" />
          </div>
        )}
      </CardContent>
    </Card>
  );
}
