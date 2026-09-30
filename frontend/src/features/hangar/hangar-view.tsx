"use client";

import { useAuth } from "@/features/auth/session/use-auth";
import { FleetCard } from "@/features/hangar/fleet-card";
import { OnboardingChecklist } from "@/features/hangar/onboarding-checklist";
import { BalanceOverviewCard } from "@/features/wallet/balance-overview-card";

export function HangarView() {
  const { user } = useAuth();

  return (
    <div className="space-y-6">
      <div>
        <p className="text-xs font-semibold tracking-[0.2em] text-highlight uppercase">Hangar</p>
        <h1 className="mt-1 text-2xl font-semibold tracking-tight text-foreground">
          Welcome back{user ? `, ${user.username}` : ""}
        </h1>
      </div>
      <div className="grid gap-6 lg:grid-cols-2">
        <div className="space-y-6">
          <OnboardingChecklist />
          <FleetCard />
        </div>
        <BalanceOverviewCard />
      </div>
    </div>
  );
}
