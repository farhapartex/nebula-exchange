import { NebulaLogo } from "@/components/brand/nebula-logo";
import { PageContainer } from "@/components/layout/page-container";

export function SiteHeader() {
  return (
    <header className="sticky top-0 z-40 border-b border-border/70 bg-background/80 backdrop-blur-md">
      <PageContainer className="flex h-16 items-center justify-between">
        <NebulaLogo />
        <span className="rounded-full border border-border-strong px-2.5 py-1 text-xs font-medium text-muted">
          Testnet
        </span>
      </PageContainer>
    </header>
  );
}
