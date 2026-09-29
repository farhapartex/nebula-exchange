"use client";

import Link from "next/link";
import { ShieldAlert, Sparkles } from "lucide-react";

import { PageContainer } from "@/components/layout/page-container";
import { Button } from "@/components/ui/button";
import { entryFeePath } from "@/features/auth/session/game-route-access";
import { useAuth } from "@/features/auth/session/use-auth";

export function AccountStateBanner() {
  const { user } = useAuth();

  if (user?.status === "PENDING_PAYMENT") {
    return (
      <div className="border-b border-accent/30 bg-accent/10">
        <PageContainer className="flex flex-wrap items-center justify-between gap-3 py-2.5">
          <p className="flex items-center gap-2 text-sm text-foreground">
            <Sparkles className="size-4 shrink-0 text-accent-soft" />
            Pay the 5 NC entry fee to launch your first mission and start trading.
          </p>
          <Button asChild size="sm">
            <Link href={entryFeePath}>Pay entry fee</Link>
          </Button>
        </PageContainer>
      </div>
    );
  }

  if (user?.status === "FROZEN") {
    return (
      <div role="alert" className="border-b border-down/40 bg-down-soft/30">
        <PageContainer className="flex items-center gap-2 py-2.5 text-sm text-foreground">
          <ShieldAlert className="size-4 shrink-0 text-down" />
          Your account is frozen. You can look around, but trading, payments and withdrawals are paused. Contact support
          to resolve it.
        </PageContainer>
      </div>
    );
  }

  return null;
}
