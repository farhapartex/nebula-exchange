import type { ReactNode } from "react";

import { GameTopBar } from "@/components/app-shell/game-top-bar";
import { RequireAuthentication } from "@/features/auth/session/require-authentication";

export default function GameLayout({ children }: { children: ReactNode }) {
  return (
    <RequireAuthentication>
      <GameTopBar />
      <main className="flex-1">{children}</main>
    </RequireAuthentication>
  );
}
