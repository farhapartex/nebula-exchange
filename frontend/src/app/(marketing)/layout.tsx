import type { ReactNode } from "react";

import { SiteHeader } from "@/components/site/site-header";

export default function MarketingLayout({ children }: { children: ReactNode }) {
  return (
    <>
      <SiteHeader />
      <main className="flex-1">{children}</main>
    </>
  );
}
