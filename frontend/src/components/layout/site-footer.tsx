import { PageContainer } from "@/components/layout/page-container";

export function SiteFooter() {
  const currentYear = new Date().getFullYear();

  return (
    <footer className="border-t border-border/70">
      <PageContainer className="flex h-14 items-center justify-between text-xs text-subtle">
        <span>© {currentYear} Nebula Exchange</span>
        <span>Base Sepolia testnet</span>
      </PageContainer>
    </footer>
  );
}
