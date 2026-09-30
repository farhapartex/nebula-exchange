import type { Metadata } from "next";

import { navigationLinkFor } from "@/components/app-shell/navigation-links";
import { PageContainer } from "@/components/layout/page-container";
import { SectionHeader } from "@/components/layout/section-header";
import { WorkshopView } from "@/features/workshop/workshop-view";

export const metadata: Metadata = {
  title: "Workshop",
};

export default function WorkshopPage() {
  const navigationLink = navigationLinkFor("workshop");
  return (
    <PageContainer className="py-8 sm:py-10">
      <SectionHeader title={navigationLink.label} description={navigationLink.description} icon={navigationLink.icon} />
      <WorkshopView />
    </PageContainer>
  );
}
