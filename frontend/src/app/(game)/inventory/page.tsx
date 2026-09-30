import type { Metadata } from "next";

import { navigationLinkFor } from "@/components/app-shell/navigation-links";
import { PageContainer } from "@/components/layout/page-container";
import { SectionHeader } from "@/components/layout/section-header";
import { InventoryView } from "@/features/inventory/inventory-view";

export const metadata: Metadata = {
  title: "Inventory",
};

export default function InventoryPage() {
  const navigationLink = navigationLinkFor("inventory");
  return (
    <PageContainer className="py-8 sm:py-10">
      <SectionHeader title={navigationLink.label} description={navigationLink.description} icon={navigationLink.icon} />
      <InventoryView />
    </PageContainer>
  );
}
