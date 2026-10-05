import type { ReactNode } from "react";

import { GameTopBar } from "@/components/app-shell/game-top-bar";

export default function GameLayout({ children }: { children: ReactNode }) {
  return (
    <>
      <GameTopBar />
      <main className="flex-1">{children}</main>
    </>
  );
}
