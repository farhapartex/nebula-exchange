"use client";

import { useState } from "react";
import Link from "next/link";
import { Coins, Lock, Wallet } from "lucide-react";
import { Popover } from "radix-ui";

import { NcAmount } from "@/components/money/nc-amount";
import { Skeleton } from "@/components/ui/skeleton";
import { useAuth } from "@/features/auth/session/use-auth";
import { BucketBalanceList } from "@/features/balances/bucket-balance-list";
import { useBalances } from "@/features/balances/use-balances";
import { cn } from "@/utils/class-names";

export function BalanceChip() {
  const { status } = useAuth();
  const balancesQuery = useBalances();
  const [isOpen, setIsOpen] = useState(false);

  if (status !== "authenticated") {
    return null;
  }
  if (!balancesQuery.data) {
    return <Skeleton className="h-9 w-28 rounded-lg" />;
  }

  const balanceSummary = balancesQuery.data;
  const hasHeldBalance = BigInt(balanceSummary.held) > 0n;

  return (
    <Popover.Root open={isOpen} onOpenChange={setIsOpen}>
      <Popover.Trigger
        aria-label="NC balance breakdown"
        className="flex h-9 items-center gap-2 rounded-lg border border-border-strong bg-surface-raised px-3 text-sm text-foreground transition-colors hover:border-accent/60 focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none data-[state=open]:border-accent/60"
      >
        <Coins className="size-4 text-warning" aria-hidden="true" />
        <NcAmount amount={balanceSummary.total} />
      </Popover.Trigger>
      <Popover.Portal>
        <Popover.Content
          align="end"
          sideOffset={8}
          collisionPadding={16}
          onOpenAutoFocus={(focusEvent) => focusEvent.preventDefault()}
          className="z-50 w-[min(22rem,calc(100vw-2rem))] rounded-xl border border-border-strong bg-surface-raised shadow-xl shadow-black/50 data-[state=open]:animate-pop-in"
        >
          <div className="border-b border-border px-4 py-3">
            <p className="text-xs text-muted">Total balance</p>
            <p className="mt-0.5 text-xl font-semibold text-foreground">
              <NcAmount amount={balanceSummary.total} />
            </p>
          </div>

          <BucketBalanceList buckets={balanceSummary.buckets} className="p-1.5" />

          <dl className="space-y-1.5 border-t border-border px-4 py-3 text-sm">
            <div className="flex items-center justify-between gap-3">
              <dt className="flex items-center gap-1.5 text-muted">
                <Lock className="size-3.5" aria-hidden="true" />
                Held by orders and bids
              </dt>
              <dd className={cn(hasHeldBalance ? "text-foreground" : "text-subtle")}>
                <NcAmount amount={balanceSummary.held} />
              </dd>
            </div>
            <div className="flex items-center justify-between gap-3">
              <dt className="flex items-center gap-1.5 text-muted">
                <Wallet className="size-3.5" aria-hidden="true" />
                Withdrawable now
              </dt>
              <dd className="text-up">
                <NcAmount amount={balanceSummary.withdrawable} />
              </dd>
            </div>
          </dl>

          <div className="border-t border-border p-1.5">
            <Link
              href="/wallet"
              onClick={() => setIsOpen(false)}
              className="flex h-9 items-center justify-center rounded-lg text-sm font-medium text-muted transition-colors hover:bg-border hover:text-foreground"
            >
              Open wallet
            </Link>
          </div>
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
  );
}
