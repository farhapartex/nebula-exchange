import type { Metadata } from "next";

import { navigationLinkFor } from "@/components/app-shell/navigation-links";
import { PageContainer } from "@/components/layout/page-container";
import { SectionHeader } from "@/components/layout/section-header";
import { BalanceOverviewCard } from "@/features/wallet/balance-overview-card";
import { TopUpPanel } from "@/features/wallet/topup-panel";
import { WalletLinkPlaceholder } from "@/features/wallet/wallet-link-placeholder";

export const metadata: Metadata = {
  title: "Wallet",
};

export default function WalletPage() {
  const navigationLink = navigationLinkFor("wallet");
  return (
    <PageContainer className="py-8 sm:py-10">
      <SectionHeader title={navigationLink.label} description={navigationLink.description} icon={navigationLink.icon} />
      <div className="grid gap-6 lg:grid-cols-2">
        <div className="space-y-6">
          <BalanceOverviewCard />
          <WalletLinkPlaceholder />
        </div>
        <TopUpPanel />
      </div>
    </PageContainer>
  );
}
