import { Gift } from "lucide-react";

import { PageContainer } from "@/components/layout/page-container";
import { NcAmount } from "@/components/money/nc-amount";
import { StatusBadge } from "@/components/ui/status-badge";
import { entryFeeInMicroUnits, starterPackEntries } from "@/features/onboarding/starter-pack";

export function EntryFeePlaceholder() {
  return (
    <PageContainer className="max-w-3xl py-8 sm:py-10">
      <h1 className="text-2xl font-semibold tracking-tight text-foreground">Pay the entry fee</h1>
      <p className="mt-1 text-sm text-muted">
        One payment of <NcAmount amount={entryFeeInMicroUnits} className="text-foreground" /> unlocks the game and your
        starter pack.
      </p>
      <div className="mt-6 rounded-2xl border border-border bg-surface/70 p-5">
        <p className="mb-4 flex items-center gap-2 text-sm font-medium text-foreground">
          <Gift className="size-4 text-accent-soft" />
          Your starter pack
        </p>
        <ul className="grid gap-3 sm:grid-cols-2">
          {starterPackEntries.map((starterPackEntry) => (
            <li key={starterPackEntry.label} className="rounded-xl border border-border bg-background/60 p-3">
              <p className="text-sm text-foreground">
                <span className="font-mono text-accent-soft">{starterPackEntry.quantity}×</span>{" "}
                {starterPackEntry.label}
              </p>
              <p className="mt-0.5 text-xs text-muted">{starterPackEntry.description}</p>
            </li>
          ))}
        </ul>
      </div>
      <StatusBadge className="mt-6" label="Card payment arrives in T-033" tone="accent" />
    </PageContainer>
  );
}
