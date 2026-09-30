import type { Metadata } from "next";

import { navigationLinkFor } from "@/components/app-shell/navigation-links";
import { PageContainer } from "@/components/layout/page-container";
import { SectionHeader } from "@/components/layout/section-header";
import { HistoryTabs } from "@/features/history/history-tabs";

export const metadata: Metadata = {
  title: "History",
};

export default function HistoryPage() {
  const navigationLink = navigationLinkFor("history");
  return (
    <PageContainer className="py-8 sm:py-10">
      <SectionHeader title={navigationLink.label} description={navigationLink.description} icon={navigationLink.icon} />
      <HistoryTabs />
    </PageContainer>
  );
}
