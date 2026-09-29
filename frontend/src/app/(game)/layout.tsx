import type { ReactNode } from "react";

import { AppTopNav } from "@/components/app-shell/app-top-nav";

export default function GameLayout({ children }: { children: ReactNode }) {
  return (
    <>
      <AppTopNav />
      <main className="flex-1">{children}</main>
    </>
  );
}
