"use client";

import { useEffect, type ReactNode } from "react";
import { usePathname, useRouter } from "next/navigation";

import { PageContainer } from "@/components/layout/page-container";
import { Skeleton } from "@/components/ui/skeleton";
import { decideGameRouteAccess } from "@/features/auth/session/game-route-access";
import { useAuth } from "@/features/auth/session/use-auth";

export function GameAccessGuard({ children }: { children: ReactNode }) {
  const router = useRouter();
  const currentPathname = usePathname();
  const { status, user } = useAuth();
  const accessDecision = decideGameRouteAccess(currentPathname, status, user?.status ?? null);
  const redirectDestination = accessDecision.kind === "redirect" ? accessDecision.destination : null;

  useEffect(() => {
    if (redirectDestination) {
      router.replace(redirectDestination);
    }
  }, [redirectDestination, router]);

  if (accessDecision.kind !== "allow") {
    return (
      <PageContainer className="space-y-4 py-10" aria-busy="true">
        <Skeleton className="h-10 w-64" />
        <Skeleton className="h-4 w-96 max-w-full" />
        <Skeleton className="h-64 w-full rounded-2xl" />
      </PageContainer>
    );
  }
  return children;
}
