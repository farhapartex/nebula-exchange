"use client";

import Link from "next/link";

import { GameLogo } from "@/components/brand/game-logo";
import { PageContainer } from "@/components/layout/page-container";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/features/auth/session/use-auth";
import { useLogOutAction } from "@/features/auth/session/use-log-out-action";

export function SiteHeader() {
  const { status, user } = useAuth();
  const { logOutAndLeave, isLoggingOut } = useLogOutAction();

  return (
    <header className="sticky top-0 z-40 border-b border-border/70 bg-background/80 backdrop-blur-md">
      <PageContainer className="flex h-16 items-center gap-6">
        <GameLogo />
        <div className="ml-auto flex items-center gap-2">
          {status === "authenticated" && user ? (
            <>
              <span className="hidden text-sm text-muted sm:inline">
                Signed in as <span className="text-foreground">{user.username}</span>
              </span>
              <Button variant="ghost" size="sm" isLoading={isLoggingOut} onClick={() => void logOutAndLeave()}>
                Log out
              </Button>
            </>
          ) : (
            <>
              <Button asChild variant="ghost" size="sm">
                <Link href="/login">Log in</Link>
              </Button>
              <Button asChild size="sm">
                <Link href="/signup">Sign up</Link>
              </Button>
            </>
          )}
        </div>
      </PageContainer>
    </header>
  );
}
