import type { Metadata } from "next";

import { navigationLinkFor } from "@/components/app-shell/navigation-links";
import { PageContainer } from "@/components/layout/page-container";
import { SectionHeader } from "@/components/layout/section-header";
import { MissionsView } from "@/features/missions/missions-view";

export const metadata: Metadata = {
  title: "Missions",
};

export default function MissionsPage() {
  const navigationLink = navigationLinkFor("missions");
  return (
    <PageContainer className="py-8 sm:py-10">
      <SectionHeader title={navigationLink.label} description={navigationLink.description} icon={navigationLink.icon} />
      <MissionsView />
    </PageContainer>
  );
}
