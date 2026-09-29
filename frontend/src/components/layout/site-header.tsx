import Link from "next/link";

import { NebulaLogo } from "@/components/brand/nebula-logo";
import { PageContainer } from "@/components/layout/page-container";
import { Button } from "@/components/ui/button";

export function SiteHeader() {
  return (
    <header className="sticky top-0 z-40 border-b border-border/70 bg-background/80 backdrop-blur-md">
      <PageContainer className="flex h-16 items-center justify-between gap-4">
        <NebulaLogo wordmarkClassName="hidden sm:inline" />
        <div className="flex items-center gap-2 sm:gap-3">
          <span className="hidden rounded-full border border-border-strong px-2.5 py-1 text-xs font-medium text-muted sm:inline">
            Testnet
          </span>
          <Button asChild variant="ghost" size="sm">
            <Link href="/login">Log in</Link>
          </Button>
          <Button asChild size="sm">
            <Link href="/signup">Sign up</Link>
          </Button>
        </div>
      </PageContainer>
    </header>
  );
}
