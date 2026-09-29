import type { ReactNode } from "react";

import { AccountStateBanner } from "@/components/app-shell/account-state-banner";
import { AppTopNav } from "@/components/app-shell/app-top-nav";
import { GameAccessGuard } from "@/components/app-shell/game-access-guard";

export default function GameLayout({ children }: { children: ReactNode }) {
  return (
    <>
      <AppTopNav />
      <AccountStateBanner />
      <main className="flex-1">
        <GameAccessGuard>{children}</GameAccessGuard>
      </main>
    </>
  );
}
