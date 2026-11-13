"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { CoinBalanceButton } from "@/components/app-shell/coin-balance-button";
import { ConnectedWalletBadge } from "@/components/app-shell/connected-wallet-badge";
import { ProfileMenu } from "@/components/app-shell/profile-menu";
import { gameNavigationLinks } from "@/components/app-shell/game-navigation";
import { GameLogo } from "@/components/brand/game-logo";
import { PageContainer } from "@/components/layout/page-container";
import { cn } from "@/utils/class-names";

export function GameTopBar() {
  const pathname = usePathname();

  return (
    <header className="sticky top-0 z-40 border-b border-border/70 bg-background/85 backdrop-blur-md">
      <PageContainer className="flex h-16 items-center gap-6">
        <GameLogo href="/fight" />
        <nav aria-label="Game" className="hidden items-center gap-1 md:flex">
          {gameNavigationLinks.map((navigationLink) => {
            const isActive = pathname.startsWith(navigationLink.href);
            if (!navigationLink.isAvailable) {
              return (
                <span
                  key={navigationLink.href}
                  title="Coming soon"
                  className="flex cursor-not-allowed items-center gap-1.5 rounded-lg px-3 py-2 text-sm text-subtle"
                >
                  {navigationLink.label}
                  <span className="rounded bg-surface-raised px-1 text-[0.625rem] text-subtle uppercase">Soon</span>
                </span>
              );
            }
            return (
              <Link
                key={navigationLink.href}
                href={navigationLink.href}
                aria-current={isActive ? "page" : undefined}
                className={cn(
                  "rounded-lg px-3 py-2 text-sm transition-colors",
                  isActive ? "bg-accent/10 text-accent-soft" : "text-muted hover:text-foreground",
                )}
              >
                {navigationLink.label}
              </Link>
            );
          })}
        </nav>
        <div className="ml-auto flex items-center gap-3">
          <ConnectedWalletBadge />
          <CoinBalanceButton />
          <ProfileMenu />
        </div>
      </PageContainer>
    </header>
  );
}
