import { Suspense, type ReactNode } from "react";

import { GameTopBar } from "@/components/app-shell/game-top-bar";
import { RequireAuthentication } from "@/features/auth/session/require-authentication";
import { CoinCheckoutReturnNotice } from "@/features/coin-packs/coin-checkout-return-notice";
import { CoinPackDialogProvider } from "@/features/coin-packs/coin-pack-dialog-context";

export default function GameLayout({ children }: { children: ReactNode }) {
  return (
    <RequireAuthentication>
      <CoinPackDialogProvider>
        <GameTopBar />
        <main className="flex-1">{children}</main>
        <Suspense>
          <CoinCheckoutReturnNotice />
        </Suspense>
      </CoinPackDialogProvider>
    </RequireAuthentication>
  );
}
